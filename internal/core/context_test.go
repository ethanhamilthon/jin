package core

import (
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/provider"
	"jin/internal/tools"
)

func contextInput(t *testing.T) PromptInput {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "AGENTS.md"), "project rules")
	writeFile(t, filepath.Join(filepath.Dir(dir), "AGENTS.md"), "parent rules")
	return PromptInput{System: "You are jin.", Dir: dir, SessionID: "s1", ToolNames: tools.Catalog(), Hooks: []string{"First hook.", "Second hook."}}
}

func TestPartsJoinToTheExactPrompt(t *testing.T) {
	in := contextInput(t)
	system := strings.TrimSpace(in.System)
	want := strings.Join([]string{
		system, strings.TrimSpace(docsPrompt), strings.TrimSpace(asyncPrompt), "First hook.", "Second hook.",
		"AGENTS.md:\n" + renderContext(ContextFiles(in.Dir)), provider.CacheBreak, sessionTail(in, system),
	}, "\n\n")
	if got := joinParts(SystemPromptParts(in)); got != want {
		t.Fatalf("parts do not join to the prompt:\n%q\nwant\n%q", got, want)
	}
	if BuildSystemPrompt(in) != want {
		t.Fatal("BuildSystemPrompt changed its output")
	}
}

func TestPartsAreLabeled(t *testing.T) {
	in := contextInput(t)
	var names []string
	for _, part := range SystemPromptParts(in) {
		names = append(names, part.Name)
	}
	got := strings.Join(names, "|")
	for _, want := range []string{"system text|jin docs|async instructions|hook 1|hook 2|AGENTS.md ", "AGENTS.md " + filepath.Join(in.Dir, "AGENTS.md"), "|cache break|environment"} {
		if !strings.Contains(got, want) {
			t.Errorf("names %q lack %q", got, want)
		}
	}
}

func TestExplainPromptFindsTheParts(t *testing.T) {
	in := contextInput(t)
	prompt := joinParts(SystemPromptParts(in)) // not built, so not remembered
	hooks := []PromptPart{{"hook a", "First hook."}, {"hook b", "Second hook."}}
	got := ExplainPrompt(prompt, in.Dir, hooks)
	if joinParts(got) != prompt {
		t.Fatalf("explained parts do not join to the prompt")
	}
	want := SystemPromptParts(in)
	if len(got) != len(want) {
		t.Fatalf("got %d parts, want %d", len(got), len(want))
	}
	if got[3].Name != "hook a" {
		t.Errorf("hook part = %q", got[3].Name)
	}
}

func TestExplainPromptKeepsWhatItCannotMatch(t *testing.T) {
	in := contextInput(t)
	prompt := joinParts(SystemPromptParts(in)) // not built, so not remembered
	got := ExplainPrompt(prompt, in.Dir, []PromptPart{{"hook a", "rendered differently"}})
	if joinParts(got) != prompt {
		t.Fatal("parts do not join to the prompt")
	}
	if !strings.HasPrefix(got[3].Name, "hooks (") || !strings.Contains(got[3].Text, "Second hook.") {
		t.Errorf("unmatched hooks part = %+v", got[3])
	}
	writeFile(t, filepath.Join(in.Dir, "AGENTS.md"), "changed")
	for _, part := range ExplainPrompt(prompt, in.Dir, nil) {
		if strings.HasPrefix(part.Name, "AGENTS.md (changed") {
			return
		}
	}
	t.Error("changed AGENTS.md not reported")
}

func TestLargestToolResults(t *testing.T) {
	call := func(id, name string) provider.Message {
		c := provider.ToolCall{ID: id}
		c.Function.Name = name
		return provider.Message{Role: "assistant", ToolCalls: []provider.ToolCall{c}}
	}
	messages := []provider.Message{
		call("1", "read"), {Role: "tool", ToolCallID: "1", Content: "ab"},
		call("2", "bash"), {Role: "tool", ToolCallID: "2", Content: "abcdef"},
		call("3", "grep"), {Role: "tool", ToolCallID: "3", Content: "abcd"},
	}
	got := LargestToolResults(messages, 2)
	if len(got) != 2 || got[0].Call.Function.Name != "bash" || got[1].Bytes != 4 {
		t.Fatalf("got %+v", got)
	}
}
