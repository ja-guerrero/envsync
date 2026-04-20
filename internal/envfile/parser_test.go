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
			errContains: "expected key=value",
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
			errContains: "mismatched double quotes",
		},
		{
			name:        "mismatched single quotes",
			input:       "FOO='hello",
			wantErr:     true,
			errContains: "mismatched single quotes",
		},
		{
			name:        "lone double quote",
			input:       `FOO="`,
			wantErr:     true,
			errContains: "unterminated quoted value",
		},
		{
			name:        "lone single quote",
			input:       "FOO='",
			wantErr:     true,
			errContains: "unterminated quoted value",
		},
		{
			name:        "whitespace inside key rejected",
			input:       "F OO=bar",
			wantErr:     true,
			errContains: "contains whitespace",
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
