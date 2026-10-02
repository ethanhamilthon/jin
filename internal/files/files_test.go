package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/clipperhouse/displaywidth"
)

func TestFind(t *testing.T) {
	for _, tc := range []struct {
		text   string
		cursor int
		raw    string
		quoted bool
		ok     bool
	}{
		{"x @foo.md", 9, "foo.md", false, true}, {`@"two words`, 11, "two words", true, true}, {"x@no", 4, "", false, false}, {"@one two", 6, "", false, false},
	} {
		got, ok := Find(tc.text, tc.cursor)
		if ok != tc.ok || ok && (got.Raw != tc.raw || got.Quoted != tc.quoted) {
			t.Fatalf("Find(%q): %#v %v", tc.text, got, ok)
		}
	}
}

func TestCandidates(t *testing.T) {
	d := t.TempDir()
	_ = os.Mkdir(filepath.Join(d, "dir"), 0755)
	_ = os.WriteFile(filepath.Join(d, "file"), nil, 0644)
	_ = os.WriteFile(filepath.Join(d, ".hidden"), nil, 0644)
	got := Candidates("", d, d, 10)
	if len(got) != 2 || !got[0].IsDir || got[1].Name != "file" {
		t.Fatalf("Candidates: %#v", got)
	}
	if got := Candidates("missing/x", d, d, 4); len(got) != 0 {
		t.Fatalf("missing dir: %#v", got)
	}
}

func TestInsert(t *testing.T) {
	if got := Insert(Candidate{Path: "two words"}); got != `"two words"` {
		t.Fatal(got)
	}
}
func TestResolve(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "é")
	_ = os.WriteFile(p, nil, 0644)
	got, ok := Resolve("é", d, d)
	if !ok || got != p {
		t.Fatalf("%q %v", got, ok)
	}
	if _, ok := Resolve("absent", d, d); ok {
		t.Fatal("missing resolved")
	}
}
func TestBlock(t *testing.T) {
	if got := Block(nil); got != "" {
		t.Fatal(got)
	}
	got := Block([]string{"/a&b"})
	want := "<attached-files>\n<file path=\"/a&amp;b\"/>\n</attached-files>"
	if got != want {
		t.Fatalf("%q", got)
	}
}
func TestExtract(t *testing.T) {
	d := t.TempDir()
	_ = os.WriteFile(filepath.Join(d, "a b"), nil, 0644)
	got, paths := Extract(`see @"a b"`, d, d)
	if got != `see `+filepath.Join(d, "a b") || len(paths) != 1 {
		t.Fatalf("%q %#v", got, paths)
	}
}
func TestShorten(t *testing.T) {
	got := Shorten("/very/long/directory/name.txt", 12)
	if displaywidth.String(got) > 12 || !strings.HasSuffix(got, "name.txt") {
		t.Fatalf("%q", got)
	}
}

func TestCandidatesKeepTheTypedDirectoryForm(t *testing.T) {
	home := t.TempDir()
	cwd := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "docs", "my plan.md"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(cwd, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cwd, "src", "app.go"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct{ raw, want string }{
		{"sr", "src/"},
		{"src", "src/"},
		{"src/", "src/app.go"},
		{"./src/a", "./src/app.go"},
		{"~/do", "~/docs/"},
		{"~/docs/", "~/docs/my plan.md"},
		{"~/docs/MY", "~/docs/my plan.md"},
	}
	for _, tc := range cases {
		got := Candidates(tc.raw, home, cwd, 10)
		if len(got) != 1 || got[0].Path != tc.want {
			t.Errorf("Candidates(%q) = %+v, want one result %q", tc.raw, got, tc.want)
		}
	}
	if got := Insert(Candidate{Path: "~/docs/my plan.md"}); got != `"~/docs/my plan.md"` {
		t.Errorf("Insert = %s", got)
	}
}
