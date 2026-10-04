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
	out := render(t, Input{ToolNames: []string{"read"}, WithPrompts: true})
	if len(out.Warnings) != 0 {
		t.Fatalf("warnings: %v", out.Warnings)
	}
}

func TestUnreadableCustomFilesWarnWithTheirPath(t *testing.T) {
	root := setup(t)
	unreadable(t, filepath.Join(root, "system-prompt.md"))
	unreadable(t, filepath.Join(root, "hooks", "bad.md"))
	unreadable(t, filepath.Join(root, "prompts", "broken.md"))
	write(t, filepath.Join(root, "hooks", "good.md"), "good hook")
	project := t.TempDir()
	unreadable(t, filepath.Join(project, ".jin", "hooks", "team.md"))
	out := render(t, Input{Dir: project, ToolNames: []string{"read"}, WithPrompts: true, ProjectHooks: true})
	for _, path := range []string{"system-prompt.md", filepath.Join("hooks", "bad.md"), filepath.Join("prompts", "broken.md"), filepath.Join(".jin", "hooks", "team.md")} {
		found := false
		for _, w := range out.Warnings {
			found = found || strings.Contains(w, path)
		}
		if !found {
			t.Errorf("no warning for %s: %v", path, out.Warnings)
		}
	}
	if !strings.Contains(out.System, "good hook") || !strings.Contains(out.System, "Tools:") {
		t.Errorf("the prompt must still be usable:\n%s", out.System)
	}
	if _, ok := out.Prompts["plan"]; !ok {
		t.Error("built-in prompts must stay")
	}
}
