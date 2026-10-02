package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/tools"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func build(t *testing.T, in PromptInput) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	if in.Dir == "" {
		in.Dir = t.TempDir()
	}
	return BuildSystemPrompt(in)
}

func TestPartsComeInOrder(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "project rules")
	prompt := BuildSystemPrompt(PromptInput{
		System: "You are jin.", Dir: dir, SessionID: "sess-1", ToolNames: tools.Catalog(), Hooks: []string{"First hook.", "Second hook."},
	})
	order := []string{"You are jin.", "Tools:\n", "First hook.", "Second hook.", "Jin documentation:", "Async tasks:", "AGENTS.md:\n", "project rules"}
	last := -1
	for _, want := range order {
		at := strings.Index(prompt, want)
		if at < 0 || at < last {
			t.Fatalf("%q is missing or out of order (at %d, after %d):\n%s", want, at, last, prompt)
		}
		last = at
	}
	if strings.Contains(prompt, "\n\n\n") {
		t.Errorf("blank lines pile up:\n%s", prompt)
	}
}

func TestDocsPointerIsAlwaysThere(t *testing.T) {
	for _, names := range [][]string{nil, {"read"}, tools.Catalog()} {
		prompt := build(t, PromptInput{System: "x", ToolNames: names})
		if !strings.Contains(prompt, "github.com/ethanhamilthon/jin/tree/main/docs") {
			t.Errorf("docs pointer missing for tools %v", names)
		}
	}
}

func TestAsyncBlockNeedsBashAndASessionID(t *testing.T) {
	cases := []struct {
		name  string
		id    string
		tools []string
		want  bool
	}{
		{"bash and id", "s1", tools.Catalog(), true},
		{"no bash", "s1", []string{"read", "edit"}, false},
		{"no id", "", tools.Catalog(), false},
		{"no tools", "s1", nil, false},
	}
	for _, c := range cases {
		prompt := build(t, PromptInput{System: "x", SessionID: c.id, ToolNames: c.tools})
		if got := strings.Contains(prompt, "jin async run"); got != c.want {
			t.Errorf("%s: async block present = %v, want %v", c.name, got, c.want)
		}
	}
	prompt := build(t, PromptInput{System: "x", SessionID: "abc-123", ToolNames: tools.Catalog()})
	if !strings.Contains(prompt, "Your session id: abc-123") {
		t.Errorf("session id missing:\n%s", prompt)
	}
}

func TestEmptyHooksAndSystemAddNothing(t *testing.T) {
	prompt := build(t, PromptInput{System: "  ", ToolNames: []string{"read"}, Hooks: []string{"", "  \n", "real hook"}})
	if !strings.HasPrefix(prompt, "Tools:") || !strings.Contains(prompt, "real hook") || strings.Contains(prompt, "\n\n\n") {
		t.Errorf("prompt:\n%s", prompt)
	}
}

func TestToolListHasOnlyGivenTools(t *testing.T) {
	prompt := build(t, PromptInput{System: "x", ToolNames: []string{"read", "todo"}})
	if !strings.Contains(prompt, "- read:") || !strings.Contains(prompt, "- todo:") || strings.Contains(prompt, "- bash:") {
		t.Fatalf("tool list wrong:\n%s", prompt)
	}
	if prompt := build(t, PromptInput{System: "x"}); !strings.Contains(prompt, "Tools: none") {
		t.Fatalf("no-tools note missing:\n%s", prompt)
	}
}

func TestAgentsMdFilesAreIncludedWithTheirSource(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := t.TempDir()
	dir := filepath.Join(root, "project")
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "project rules")
	writeFile(t, filepath.Join(root, "AGENTS.md"), "parent rules")
	prompt := BuildSystemPrompt(PromptInput{System: "x", Dir: dir})
	for _, want := range []string{"project rules", "parent rules", "parent directory, not the current project", "current project"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
}

// Commands are run only in text the user wrote for jin. An AGENTS.md comes
// from a repository, so its braces must come through untouched and nothing
// may run; building the prompt does not run commands at all.
func TestAgentsMdIsNeverExecutedOrExpanded(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	marker := filepath.Join(t.TempDir(), "ran")
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "Rules {{touch "+marker+"}} and {{pwd}} and {{tools}}")
	prompt := BuildSystemPrompt(PromptInput{System: "x", Dir: dir, ToolNames: tools.Catalog()})
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("a command in AGENTS.md was run")
	}
	if !strings.Contains(prompt, "{{touch "+marker+"}} and {{pwd}} and {{tools}}") {
		t.Errorf("AGENTS.md text was changed:\n%s", prompt)
	}
}

func TestSidePromptsFallBackToTheDefaults(t *testing.T) {
	a := NewAgent(nil, "sys", tools.NewRegistry())
	if !strings.Contains(a.compactPrompt(), "compacted") || !strings.Contains(a.handoffPrompt(), "new session") {
		t.Errorf("defaults not used:\n%s\n%s", a.compactPrompt(), a.handoffPrompt())
	}
	a.SetSidePrompts("my compact", "  ")
	if a.compactPrompt() != "my compact" || !strings.Contains(a.handoffPrompt(), "new session") {
		t.Errorf("compact = %q, handoff = %q", a.compactPrompt(), a.handoffPrompt())
	}
	a.SetSystemPrompt("new system")
	if a.systemPrompt != "new system" {
		t.Error("SetSystemPrompt did not set the prompt")
	}
}
