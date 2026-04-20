package envfile

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Write writes key=value pairs to a .env file at path.
// Keys are sorted alphabetically. Values are double-quoted when needed.
// File permissions: 0600.
func Write(path string, vars map[string]string) error {
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		v := vars[k]
		if needsQuoting(v) {
			fmt.Fprintf(&b, "%s=%s\n", k, strconv.Quote(v))
		} else {
			fmt.Fprintf(&b, "%s=%s\n", k, v)
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0600)
}

// WriteFile writes a structured File to disk, preserving ordering,
// comments, and quote styles from the original parse.
func WriteFile(path string, f *File) error {
	var b strings.Builder

	for _, line := range f.Lines {
		switch l := line.(type) {
		case *Entry:
			writeEntry(&b, l)
		case *CommentLine:
			if l.Text == "" {
				b.WriteByte('\n')
			} else {
				b.WriteString(l.Text)
				b.WriteByte('\n')
			}
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0600)
}

func writeEntry(b *strings.Builder, e *Entry) {
	b.WriteString(e.Key)
	b.WriteByte('=')

	switch e.QuoteStyle {
	case QuoteDouble:
		b.WriteString(strconv.Quote(e.Value))
	case QuoteSingle:
		b.WriteByte('\'')
		b.WriteString(e.Value)
		b.WriteByte('\'')
	default:
		b.WriteString(e.Value)
	}

	if e.InlineComment != "" {
		b.WriteString(" # ")
		b.WriteString(e.InlineComment)
	}

	b.WriteByte('\n')
}

func needsQuoting(v string) bool {
	if v == "" {
		return false
	}
	return strings.ContainsAny(v, " \t\n\r\"'#$`\\")
}
