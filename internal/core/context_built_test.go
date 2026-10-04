package core

import (
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/provider"
)

func TestExplainBuiltPromptKeepsEveryHookAndFile(t *testing.T) {
	in := contextInput(t)
	in.Hooks = []string{"Hook with output: 42\nline two", "Second hook."}
	prompt := BuildSystemPrompt(in)
	writeFile(t, filepath.Join(in.Dir, "AGENTS.md"), "edited after start")
	var names []string
	for _, part := range ExplainPrompt(prompt, in.Dir, []PromptPart{{"hook a", "Hook with {{cmd}}"}}) {
		names = append(names, part.Name)
	}
	got := strings.Join(names, "|")
	for _, want := range []string{"|hook 1|hook 2|", "AGENTS.md " + filepath.Join(in.Dir, "AGENTS.md"), "|cache break|"} {
		if !strings.Contains(got, want) {
			t.Errorf("names %q lack %q", got, want)
		}
	}
	if strings.Contains(got, "changed") || strings.Contains(got, "with command output") {
		t.Errorf("fallback used: %q", got)
	}
}

func TestEffectiveMessagesPrunesOldResultsPastTheThreshold(t *testing.T) {
	call := provider.ToolCall{ID: "1"}
	call.Function.Name = "read"
	big := strings.Repeat("x", 4000)
	messages := []provider.Message{{Role: "user", Content: "first"}, {Role: "assistant", ToolCalls: []provider.ToolCall{call}}, {Role: "tool", ToolCallID: "1", Content: big}}
	for range pruneKeepTurns {
		messages = append(messages, provider.Message{Role: "user", Content: "next"})
	}
	if got := EffectiveMessages(messages, 0, 100000); len(got[2].Content) != len(big) {
		t.Error("pruned below the threshold")
	}
	got := EffectiveMessages(messages, 0, 1200)
	if len(got[2].Content) >= len(big) || LargestToolResults(got, 1)[0].Bytes >= len(big) {
		t.Errorf("not pruned past the threshold: %d bytes", len(got[2].Content))
	}
	if len(messages[2].Content) != len(big) {
		t.Error("input changed")
	}
}
