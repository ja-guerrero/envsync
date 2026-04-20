package envfile

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name:  "simple key=value",
			input: "FOO=bar",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "double-quoted value",
			input: `FOO="hello world"`,
			want:  map[string]string{"FOO": "hello world"},
		},
		{
			name:  "single-quoted value",
			input: "FOO='hello world'",
			want:  map[string]string{"FOO": "hello world"},
		},
		{
			name:  "double-quoted escape sequences",
			input: `FOO="line1\nline2"`,
			want:  map[string]string{"FOO": "line1\nline2"},
		},
		{
			name:  "single-quoted preserves backslash",
			input: `FOO='line1\nline2'`,
			want:  map[string]string{"FOO": `line1\nline2`},
		},
		{
			name:  "export prefix",
			input: "export FOO=bar",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "export not treated as prefix when inside key",
			input: "MY_EXPORT_VAR=bar",
			want:  map[string]string{"MY_EXPORT_VAR": "bar"},
		},
		{
			name:  "empty value",
			input: "FOO=",
			want:  map[string]string{"FOO": ""},
		},
		{
			name:  "value with equals sign",
			input: "FOO=a=b=c",
			want:  map[string]string{"FOO": "a=b=c"},
		},
		{
			name:  "hash inside unquoted value is not a comment",
			input: "FOO=bar#baz",
			want:  map[string]string{"FOO": "bar#baz"},
		},
		{
			name:  "comments and blank lines ignored",
			input: "# comment\n\nFOO=bar\n# another\n",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "whitespace trimmed around key and value",
			input: "  FOO  =  bar  ",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "multiple vars",
			input: "FOO=bar\nBAZ=qux",
			want:  map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name:        "missing equals sign",
			input:       "FOOBAR",
			wantErr:     true,
			errContains: "expected KEY=VALUE",
		},
		{
			name:        "empty key",
			input:       "=value",
			wantErr:     true,
			errContains: "empty key",
		},
		{
			name:        "duplicate key",
			input:       "FOO=bar\nFOO=baz",
			wantErr:     true,
			errContains: "duplicate key",
		},
		{
			name:        "duplicate key reports correct line number",
			input:       "FOO=bar\nFOO=baz",
			wantErr:     true,
			errContains: "line 2:",
		},
		{
			name:        "empty key reports correct line number",
			input:       "FOO=bar\n=baz",
			wantErr:     true,
			errContains: "line 2:",
		},
		{
			name:        "mismatched double quotes",
			input:       `FOO="hello`,
			wantErr:     true,
			errContains: "unterminated double quote",
		},
		{
			name:        "mismatched single quotes",
			input:       "FOO='hello",
			wantErr:     true,
			errContains: "unterminated single quote",
		},
		{
			name:        "lone double quote",
			input:       `FOO="`,
			wantErr:     true,
			errContains: "unterminated double quote",
		},
		{
			name:        "lone single quote",
			input:       "FOO='",
			wantErr:     true,
			errContains: "unterminated single quote",
		},
		{
			name:        "whitespace inside key rejected",
			input:       "F OO=bar",
			wantErr:     true,
			errContains: "invalid key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantEntries []Entry
		wantErr     bool
		errContains string
	}{
		{
			name:  "simple key=value",
			input: "FOO=bar",
			wantEntries: []Entry{
				{Key: "FOO", Value: "bar", QuoteStyle: QuoteNone, Line: 1},
			},
		},
		{
			name:  "double-quoted value",
			input: `FOO="hello world"`,
			wantEntries: []Entry{
				{Key: "FOO", Value: "hello world", QuoteStyle: QuoteDouble, Line: 1},
			},
		},
		{
			name:  "single-quoted value",
			input: "FOO='hello world'",
			wantEntries: []Entry{
				{Key: "FOO", Value: "hello world", QuoteStyle: QuoteSingle, Line: 1},
			},
		},
		{
			name:  "double-quoted escapes",
			input: `FOO="line1\nline2\ttab\"\\"`,
			wantEntries: []Entry{
				{Key: "FOO", Value: "line1\nline2\ttab\"\\", QuoteStyle: QuoteDouble, Line: 1},
			},
		},
		{
			name:  "single-quoted preserves backslash",
			input: `FOO='line1\nline2'`,
			wantEntries: []Entry{
				{Key: "FOO", Value: `line1\nline2`, QuoteStyle: QuoteSingle, Line: 1},
			},
		},
		{
			name:  "inline comment with space before hash",
			input: "FOO=bar # this is a comment",
			wantEntries: []Entry{
				{Key: "FOO", Value: "bar", InlineComment: "this is a comment", QuoteStyle: QuoteNone, Line: 1},
			},
		},
		{
			name:  "hash without preceding space is part of value",
			input: "FOO=bar#notcomment",
			wantEntries: []Entry{
				{Key: "FOO", Value: "bar#notcomment", QuoteStyle: QuoteNone, Line: 1},
			},
		},
		{
			name:  "hash inside double quotes is not a comment",
			input: `FOO="bar # inside"`,
			wantEntries: []Entry{
				{Key: "FOO", Value: "bar # inside", QuoteStyle: QuoteDouble, Line: 1},
			},
		},
		{
			name:  "comment after quoted value",
			input: `FOO="bar" # comment`,
			wantEntries: []Entry{
				{Key: "FOO", Value: "bar", InlineComment: "comment", QuoteStyle: QuoteDouble, Line: 1},
			},
		},
		{
			name:  "export prefix",
			input: "export FOO=bar",
			wantEntries: []Entry{
				{Key: "FOO", Value: "bar", QuoteStyle: QuoteNone, Line: 1},
			},
		},
		{
			name:  "empty value",
			input: "FOO=",
			wantEntries: []Entry{
				{Key: "FOO", Value: "", QuoteStyle: QuoteNone, Line: 1},
			},
		},
		{
			name:  "value with equals sign",
			input: "FOO=a=b=c",
			wantEntries: []Entry{
				{Key: "FOO", Value: "a=b=c", QuoteStyle: QuoteNone, Line: 1},
			},
		},
		{
			name:  "preserves comment lines in structure",
			input: "# header\n\nFOO=bar\n# mid\nBAZ=qux",
			wantEntries: []Entry{
				{Key: "FOO", Value: "bar", QuoteStyle: QuoteNone, Line: 3},
				{Key: "BAZ", Value: "qux", QuoteStyle: QuoteNone, Line: 5},
			},
		},
		{
			name:        "key starts with digit",
			input:       "3FOO=bar",
			wantErr:     true,
			errContains: "invalid key",
		},
		{
			name:        "key with special char",
			input:       "FOO-BAR=baz",
			wantErr:     true,
			errContains: "invalid key",
		},
		{
			name:        "mismatched double quotes",
			input:       `FOO="hello`,
			wantErr:     true,
			errContains: "unterminated double quote",
		},
		{
			name:        "mismatched single quotes",
			input:       "FOO='hello",
			wantErr:     true,
			errContains: "unterminated single quote",
		},
		{
			name:        "duplicate key",
			input:       "FOO=bar\nFOO=baz",
			wantErr:     true,
			errContains: "duplicate key",
		},
		{
			name:        "empty key",
			input:       "=value",
			wantErr:     true,
			errContains: "empty key",
		},
		{
			name:        "missing equals",
			input:       "FOOBAR",
			wantErr:     true,
			errContains: "expected KEY=VALUE",
		},
		{
			name:        "invalid escape in double quotes",
			input:       `FOO="hello\x"`,
			wantErr:     true,
			errContains: "invalid escape",
		},
		{
			name:  "whitespace trimmed around key and value",
			input: "  FOO  =  bar  ",
			wantEntries: []Entry{
				{Key: "FOO", Value: "bar", QuoteStyle: QuoteNone, Line: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := ParseFile(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			entries := f.Entries()
			if len(entries) != len(tt.wantEntries) {
				t.Fatalf("got %d entries, want %d", len(entries), len(tt.wantEntries))
			}
			for i, want := range tt.wantEntries {
				got := entries[i]
				if got.Key != want.Key {
					t.Errorf("entry[%d].Key = %q, want %q", i, got.Key, want.Key)
				}
				if got.Value != want.Value {
					t.Errorf("entry[%d].Value = %q, want %q", i, got.Value, want.Value)
				}
				if got.QuoteStyle != want.QuoteStyle {
					t.Errorf("entry[%d].QuoteStyle = %d, want %d", i, got.QuoteStyle, want.QuoteStyle)
				}
				if got.Line != want.Line {
					t.Errorf("entry[%d].Line = %d, want %d", i, got.Line, want.Line)
				}
				if got.InlineComment != want.InlineComment {
					t.Errorf("entry[%d].InlineComment = %q, want %q", i, got.InlineComment, want.InlineComment)
				}
			}
		})
	}
}

func TestFileToMap(t *testing.T) {
	f := &File{
		Lines: []any{
			&CommentLine{Text: "# header", Line: 1},
			&Entry{Key: "FOO", Value: "bar", Line: 2},
			&Entry{Key: "BAZ", Value: "qux", Line: 3},
		},
	}

	m := f.ToMap()
	if len(m) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m))
	}
	if m["FOO"] != "bar" {
		t.Errorf("FOO: got %q, want %q", m["FOO"], "bar")
	}
	if m["BAZ"] != "qux" {
		t.Errorf("BAZ: got %q, want %q", m["BAZ"], "qux")
	}
}

func TestFileEntries(t *testing.T) {
	f := &File{
		Lines: []any{
			&CommentLine{Text: "# header", Line: 1},
			&Entry{Key: "A", Value: "1", Line: 2},
			&CommentLine{Text: "", Line: 3},
			&Entry{Key: "B", Value: "2", Line: 4},
		},
	}

	entries := f.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Key != "A" || entries[1].Key != "B" {
		t.Errorf("wrong entry order: %s, %s", entries[0].Key, entries[1].Key)
	}
}

func FuzzParseFile(f *testing.F) {
	// Seed corpus with representative inputs
	f.Add("FOO=bar")
	f.Add(`FOO="hello world"`)
	f.Add("FOO='hello world'")
	f.Add(`FOO="line1\nline2"`)
	f.Add("# comment\nFOO=bar\nBAZ=qux")
	f.Add("FOO=bar # comment")
	f.Add("FOO=bar#notcomment")
	f.Add("export FOO=bar")
	f.Add("FOO=")
	f.Add("FOO=a=b=c")
	f.Add("")
	f.Add("  FOO  =  bar  ")

	f.Fuzz(func(t *testing.T, input string) {
		// ParseFile must never panic. Errors are fine.
		_, _ = ParseFile(strings.NewReader(input))
	})
}
