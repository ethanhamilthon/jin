package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportWritesMarkdownIntoTheProjectFolderWithoutOverwriting(t *testing.T) {
	a, s := persistedApp(t)
	a.active = s
	a.exportSession()
	a.exportSession()
	for _, name := range []string{"title-s1.md", "title-s1-2.md"} {
		data, err := os.ReadFile(filepath.Join(a.dir, name))
		if err != nil || !strings.HasPrefix(string(data), "# title\n") {
			t.Fatalf("%s: %v, %q", name, err, data)
		}
	}
	if !strings.Contains(lastText(s), filepath.Join(a.dir, "title-s1-2.md")) {
		t.Errorf("the chat does not show the file: %q", lastText(s))
	}
}

func TestExportOfAnUnsavedSessionSaysNothingIsThere(t *testing.T) {
	a, _ := layoutApp(t)
	a.exportSession()
	if !strings.HasPrefix(lastText(a.active), "Nothing to export yet") {
		t.Fatalf("last entry %q", lastText(a.active))
	}
}

func TestExportNameIsASafeFileName(t *testing.T) {
	cases := map[string]string{"Fix: the/login?": "Fix-the-login", "../../etc": "etc", " ": "session"}
	for title, want := range cases {
		if got := exportName(title); got != want {
			t.Errorf("exportName(%q) = %q, want %q", title, got, want)
		}
	}
}
