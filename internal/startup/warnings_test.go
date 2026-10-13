package startup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func unreadable(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), path); err != nil {
		t.Fatal(err)
	}
}

func TestMissingFilesAreSilent(t *testing.T) {
	setup(t)
	out := render(t, Input{WithPrompts: true})
	if len(out.Warnings) != 0 {
		t.Fatalf("warnings: %v", out.Warnings)
	}
}

func TestUnreadablePromptWarnsButCustomSystemFileIsIgnored(t *testing.T) {
	root := setup(t)
	customPath := filepath.Join(root, "system-prompt.md")
	unreadable(t, customPath)
	unreadable(t, filepath.Join(root, "prompts", "broken.md"))
	project := t.TempDir()
	out := render(t, Input{Dir: project, WithPrompts: true})
	if _, err := os.Lstat(customPath); err != nil {
		t.Fatalf("custom file changed: %v", err)
	}
	for _, path := range []string{filepath.Join("prompts", "broken.md")} {
		found := false
		for _, w := range out.Warnings {
			found = found || strings.Contains(w, path)
		}
		if !found {
			t.Errorf("no warning for %s: %v", path, out.Warnings)
		}
	}
	if out.System == "" {
		t.Error("the prompt must still be usable")
	}
	if _, ok := out.Prompts["plan"]; !ok {
		t.Error("built-in prompts must stay")
	}
}
