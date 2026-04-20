package envfile

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func Parse(r io.Reader) (map[string]string, error) {
	scanner := bufio.NewScanner(r)
	env := make(map[string]string)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle `export ` prefix
		line = strings.TrimPrefix(line, "export ")

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: expected key=value, got %q", lineNum, line)
		}

		// Trim AFTER splitting
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNum)
		}

		// Reject whitespace in keys
		if strings.ContainsAny(key, " \t\n\r") {
			return nil, fmt.Errorf("line %d: invalid key %q (contains whitespace)", lineNum, key)
		}

		// Reject duplicate keys (FIX)
		if _, exists := env[key]; exists {
			return nil, fmt.Errorf("line %d: duplicate key %q", lineNum, key)
		}

		// Precompute quote state
		startsDouble := strings.HasPrefix(value, `"`)
		endsDouble := strings.HasSuffix(value, `"`)
		startsSingle := strings.HasPrefix(value, `'`)
		endsSingle := strings.HasSuffix(value, `'`)

		// Reject lone quote
		if len(value) == 1 && (startsDouble || startsSingle) {
			return nil, fmt.Errorf("line %d: unterminated quoted value: %q", lineNum, value)
		}

		// Validate matching quotes
		if startsDouble != endsDouble {
			return nil, fmt.Errorf("line %d: mismatched double quotes: %q", lineNum, value)
		}
		if startsSingle != endsSingle {
			return nil, fmt.Errorf("line %d: mismatched single quotes: %q", lineNum, value)
		}

		// Handle quoted values
		if len(value) >= 2 {
			switch {
			case startsDouble && endsDouble:
				unquoted, err := strconv.Unquote(value)
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid double-quoted value: %w", lineNum, err)
				}
				value = unquoted

			case startsSingle && endsSingle:
				value = value[1 : len(value)-1]
			}
		}

		env[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return env, nil
}
