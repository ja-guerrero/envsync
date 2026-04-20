package envfile

import "fmt"

// Quote represents the quoting style of a value.
type Quote int

const (
	QuoteNone   Quote = iota
	QuoteSingle
	QuoteDouble
)

// Entry represents one parsed key=value line from a .env file.
type Entry struct {
	Key           string
	Value         string
	InlineComment string // text after the comment delimiter, empty if none
	QuoteStyle    Quote
	Line          int // 1-based line number in the source file
}

// CommentLine represents a standalone comment or blank line.
type CommentLine struct {
	Text string // full text including "#" prefix, or empty for blank lines
	Line int
}

// File is the full parsed representation of a .env file.
// It preserves ordering, comments, and quote styles.
type File struct {
	// Lines stores every line in order: either *Entry or *CommentLine.
	Lines []any
}

// Entries returns only the key=value entries, in file order.
func (f *File) Entries() []*Entry {
	var out []*Entry
	for _, l := range f.Lines {
		if e, ok := l.(*Entry); ok {
			out = append(out, e)
		}
	}
	return out
}

// ToMap returns a simple key->value map (loses ordering/comments).
func (f *File) ToMap() map[string]string {
	m := make(map[string]string)
	for _, e := range f.Entries() {
		m[e.Key] = e.Value
	}
	return m
}

// ParseError provides structured error reporting with location.
type ParseError struct {
	Line int
	Col  int
	Msg  string
}

func (e *ParseError) Error() string {
	if e.Col > 0 {
		return fmt.Sprintf("line %d, col %d: %s", e.Line, e.Col, e.Msg)
	}
	return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
}
