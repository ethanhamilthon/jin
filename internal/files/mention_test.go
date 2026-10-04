package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMentionRoundTrip(t *testing.T) {
	cases := []string{
		`simple.txt`,
		`two words.md`,
		`quote"inside.txt`,
		`quote"and spaces.txt`,
		`backslash\inside.txt`,
		`both "quotes" and \backslashes\ test.go`,
		`trailing\`,
		`"wrapped"`,
	}

	for _, path := range cases {
		inserted := Insert(Candidate{Path: path})
		text := "prefix @" + inserted + " suffix"
		atIdx := strings.Index(text, "@")

		// Cursor inside token
		for offset := 1; offset <= len(inserted); offset++ {
			cursor := atIdx + offset
			tok, ok := Find(text, cursor)
			if !ok {
				t.Fatalf("Find failed for path %q at cursor %d, text: %q", path, cursor, text)
			}
			if tok.Path != path {
				t.Errorf("roundtrip mismatch for %q: got %q, want %q", path, tok.Path, path)
			}
		}
	}
}

func TestUnclosedEscapedMention(t *testing.T) {
	text := `@"foo\"bar`
	tok, ok := Find(text, len(text))
	if !ok {
		t.Fatal("Find should match unclosed mention")
	}
	if tok.Path != `foo"bar` {
		t.Errorf("got %q, want %q", tok.Path, `foo"bar`)
	}
}

func TestExtractWithQuotes(t *testing.T) {
	d := t.TempDir()
	name := `file "with" quotes`
	p := filepath.Join(d, name)
	if err := os.WriteFile(p, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	inserted := Insert(Candidate{Path: name})
	text := "see @" + inserted
	got, paths := Extract(text, d, d)
	if len(paths) != 1 || paths[0] != p {
		t.Fatalf("Extract paths = %v, want [%q]", paths, p)
	}
	if got != "see "+p {
		t.Fatalf("Extract text = %q, want %q", got, "see "+p)
	}
}
