package envfile

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// keyRegex matches valid env var names: start with letter or underscore,
// followed by letters, digits, or underscores.
var keyRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ParseFile parses an env file and returns a structured *File that preserves
// comments, blank lines, quote styles, and inline comments.
func ParseFile(r io.Reader) (*File, error) {
	scanner := bufio.NewScanner(r)
	f := &File{}
	seen := make(map[string]int) // key -> line number of first occurrence
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)

		// Blank lines and comment lines are preserved as CommentLine.
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			f.Lines = append(f.Lines, &CommentLine{Text: trimmed, Line: lineNum})
			continue
		}

		// Strip optional "export " prefix.
		line := trimmed
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			line = strings.TrimSpace(line)
		}

		// Split on first '='.
		eqIdx := strings.IndexByte(line, '=')
		if eqIdx < 0 {
			return nil, &ParseError{Line: lineNum, Msg: "expected KEY=VALUE, got " + line}
		}

		key := strings.TrimSpace(line[:eqIdx])
		rest := line[eqIdx+1:] // everything after '='

		// Validate key.
		if key == "" {
			return nil, &ParseError{Line: lineNum, Msg: "empty key"}
		}
		if !keyRegex.MatchString(key) {
			return nil, &ParseError{Line: lineNum, Msg: "invalid key " + key}
		}

		// Duplicate key check.
		if prev, exists := seen[key]; exists {
			return nil, &ParseError{Line: lineNum, Msg: "duplicate key " + key + " (first seen on line " + itoa(prev) + ")"}
		}
		seen[key] = lineNum

		// Parse value.
		value, comment, style, err := parseValue(rest, lineNum)
		if err != nil {
			return nil, err
		}

		f.Lines = append(f.Lines, &Entry{
			Key:           key,
			Value:         value,
			InlineComment: comment,
			QuoteStyle:    style,
			Line:          lineNum,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return f, nil
}

// Parse is the backward-compatible wrapper that returns a flat map.
func Parse(r io.Reader) (map[string]string, error) {
	f, err := ParseFile(r)
	if err != nil {
		return nil, err
	}
	return f.ToMap(), nil
}

// parseValue dispatches to the correct value parser based on quoting.
func parseValue(s string, lineNum int) (value, comment string, style Quote, err error) {
	s = strings.TrimSpace(s)

	switch {
	case strings.HasPrefix(s, `"`):
		value, comment, err = parseDoubleQuoted(s, lineNum)
		style = QuoteDouble
	case strings.HasPrefix(s, `'`):
		value, comment, err = parseSingleQuoted(s, lineNum)
		style = QuoteSingle
	default:
		value, comment = parseRaw(s)
		style = QuoteNone
	}
	return
}

// parseDoubleQuoted parses a double-quoted value, handling escape sequences.
// Only \n, \t, \", and \\ are valid escapes; anything else is an error.
func parseDoubleQuoted(s string, lineNum int) (value, comment string, err error) {
	if len(s) < 2 {
		return "", "", &ParseError{Line: lineNum, Msg: "unterminated double quote"}
	}

	// Find the closing quote by scanning character by character.
	var b strings.Builder
	i := 1 // skip opening quote
	for i < len(s) {
		ch := s[i]
		if ch == '\\' {
			if i+1 >= len(s) {
				return "", "", &ParseError{Line: lineNum, Msg: "unterminated double quote"}
			}
			next := s[i+1]
			switch next {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			default:
				return "", "", &ParseError{Line: lineNum, Msg: "invalid escape \\" + string(next) + " in double-quoted value"}
			}
			i += 2
			continue
		}
		if ch == '"' {
			// Closing quote found.
			after := strings.TrimSpace(s[i+1:])
			comment = parseTrailingComment(after)
			return b.String(), comment, nil
		}
		b.WriteByte(ch)
		i++
	}

	return "", "", &ParseError{Line: lineNum, Msg: "unterminated double quote"}
}

// parseSingleQuoted parses a single-quoted value. No escape processing.
func parseSingleQuoted(s string, lineNum int) (value, comment string, err error) {
	if len(s) < 2 {
		return "", "", &ParseError{Line: lineNum, Msg: "unterminated single quote"}
	}

	// Find the closing single quote (no escapes in single-quoted strings).
	closeIdx := strings.Index(s[1:], "'")
	if closeIdx < 0 {
		return "", "", &ParseError{Line: lineNum, Msg: "unterminated single quote"}
	}
	closeIdx++ // adjust for the offset of s[1:]

	value = s[1:closeIdx]
	after := strings.TrimSpace(s[closeIdx+1:])
	comment = parseTrailingComment(after)
	return value, comment, nil
}

// parseRaw parses an unquoted value. A space followed by '#' starts a comment.
func parseRaw(s string) (value, comment string) {
	// Look for " #" pattern (space before hash) to split value from comment.
	idx := strings.Index(s, " #")
	if idx >= 0 {
		value = strings.TrimSpace(s[:idx])
		comment = strings.TrimSpace(s[idx+2:]) // skip " #"
		return
	}
	value = strings.TrimSpace(s)
	return value, ""
}

// parseTrailingComment extracts an inline comment from whatever follows
// a closing quote. Expects the input to already be TrimSpaced.
// Returns the comment text (without the '#') or empty string.
func parseTrailingComment(after string) string {
	if strings.HasPrefix(after, "#") {
		return strings.TrimSpace(after[1:])
	}
	return ""
}

// itoa converts an int to a string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
