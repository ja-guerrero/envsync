package envfile

import (
	"os"
	"strings"
	"testing"
)

func TestWriteFilePreservesStructure(t *testing.T) {
	input := "# Database config\nDB_HOST=localhost\nDB_PORT=5432\n\n# API keys\nAPI_KEY='sk_test_123'\n"

	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	tmp, err := os.CreateTemp("", "envfile-*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	if err := WriteFile(tmp.Name(), f); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != input {
		t.Errorf("round-trip mismatch:\ngot:  %q\nwant: %q", string(got), input)
	}
}

func TestWriteMapSortsKeys(t *testing.T) {
	vars := map[string]string{
		"ZEBRA": "z",
		"APPLE": "a",
		"MANGO": "m",
	}

	tmp, err := os.CreateTemp("", "envfile-*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	if err := Write(tmp.Name(), vars); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	want := "APPLE=a\nMANGO=m\nZEBRA=z\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", string(got), want)
	}
}

func TestWriteMapQuotesWhenNeeded(t *testing.T) {
	vars := map[string]string{
		"SIMPLE":   "hello",
		"SPACED":   "hello world",
		"NEWLINED": "line1\nline2",
		"HASHED":   "foo#bar",
		"EMPTY":    "",
	}

	tmp, err := os.CreateTemp("", "envfile-*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	if err := Write(tmp.Name(), vars); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	content := string(got)
	cases := []struct {
		substr string
		reason string
	}{
		{"SIMPLE=hello\n", "plain value should not be quoted"},
		{"SPACED=\"hello world\"\n", "value with space should be double-quoted"},
		{"NEWLINED=\"line1\\nline2\"\n", "value with newline should be double-quoted with escape"},
		{"HASHED=\"foo#bar\"\n", "value with hash should be double-quoted"},
		{"EMPTY=\n", "empty value should not be quoted"},
	}
	for _, c := range cases {
		if !strings.Contains(content, c.substr) {
			t.Errorf("expected %q in output (%s)\nfull output: %q", c.substr, c.reason, content)
		}
	}
}

func TestWriteFilePreservesInlineComments(t *testing.T) {
	input := "FOO=bar # important note\n"

	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	tmp, err := os.CreateTemp("", "envfile-*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	if err := WriteFile(tmp.Name(), f); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != input {
		t.Errorf("inline comment not preserved:\ngot:  %q\nwant: %q", string(got), input)
	}
}

func TestWriteFilePermissions(t *testing.T) {
	f := &File{
		Lines: []any{
			&Entry{Key: "SECRET", Value: "hunter2", QuoteStyle: QuoteNone},
		},
	}

	tmp, err := os.CreateTemp("", "envfile-*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	if err := WriteFile(tmp.Name(), f); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	info, err := os.Stat(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("file mode = %04o, want 0600", mode)
	}
}
