package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSystemPromptFillsEnvironment(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := t.TempDir()
	dir := filepath.Join(root, "project")
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "project rules")
	writeFile(t, filepath.Join(root, "AGENTS.md"), "parent rules")
	prompt := SystemPrompt(dir)
	if strings.Contains(prompt, "{{") {
		t.Fatalf("unreplaced placeholder in prompt:\n%s", prompt)
	}
	if !strings.Contains(prompt, "- Working directory: "+dir+"\n") {
		t.Fatalf("working directory missing:\n%s", prompt)
	}
	for _, want := range []string{"project rules", "parent rules", "parent directory, not the current project", "current project"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
