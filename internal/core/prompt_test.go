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
	prompt := SystemPrompt(dir, nil, false)
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

func TestSystemPromptPlacesEnabledHooksBeforeAgentsMd(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	hooksDir := filepath.Join(os.Getenv("HOME"), ".jin-dev", "hooks")
	writeFile(t, filepath.Join(hooksDir, "b.md"), "Second hook.")
	writeFile(t, filepath.Join(hooksDir, "a.md"), "First hook.")
	writeFile(t, filepath.Join(hooksDir, "off.md"), "Hidden hook.")
	prompt := SystemPrompt(t.TempDir(), []string{"off"}, false)
	first, second, agents := strings.Index(prompt, "First hook."), strings.Index(prompt, "Second hook."), strings.Index(prompt, "AGENTS.md:\n")
	if first < 0 || first > second || second > agents {
		t.Errorf("hooks should come alphabetically before the AGENTS.md block:\n%s", prompt)
	}
	if strings.Contains(prompt, "Hidden hook.") || strings.Contains(prompt, "<hook") {
		t.Errorf("disabled hook or XML in prompt:\n%s", prompt)
	}
}

func TestSystemPromptWithoutHooksHasNoGap(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if prompt := SystemPrompt(t.TempDir(), nil, false); !strings.Contains(prompt, "\n\nAGENTS.md:\n") || strings.Contains(prompt, "\n\n\n") {
		t.Errorf("unexpected blank lines:\n%s", prompt)
	}
}

func TestSystemPromptJinDocsOnlyWhenEnabled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if prompt := SystemPrompt(t.TempDir(), nil, false); strings.Contains(prompt, "Jin documentation:") {
		t.Errorf("docs pointer present while off:\n%s", prompt)
	}
	prompt := SystemPrompt(t.TempDir(), nil, true)
	docs, agents := strings.Index(prompt, "Jin documentation:"), strings.Index(prompt, "AGENTS.md:\n")
	if docs < 0 || docs > agents || !strings.Contains(prompt, "github.com/ethanhamilthon/jin/tree/main/docs") || strings.Contains(prompt, "\n\n\n") {
		t.Errorf("docs pointer misplaced:\n%s", prompt)
	}
}
