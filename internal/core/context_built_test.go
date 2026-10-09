package core

import (
	"strings"
	"testing"

	"jin/internal/provider"
)

func TestExplainBuiltPromptNamesTheParts(t *testing.T) {
	parts := []PromptPart{{"system text", "Be brief.\n\n"}, {"command: jin docs", "Pointer."}, {"system text", "\n\nEnd."}}
	prompt := BuildSystemPrompt(parts)
	if prompt != "Be brief.\n\nPointer.\n\nEnd." {
		t.Fatalf("prompt = %q", prompt)
	}
	got := ExplainPrompt(prompt)
	if len(got) != 3 || got[1].Name != "command: jin docs" || joinParts(got) != joinParts(parts) {
		t.Errorf("parts = %+v", got)
	}
	if other := ExplainPrompt("never built"); len(other) != 1 || other[0].Name != "system prompt" {
		t.Errorf("unknown prompt = %+v", other)
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
	if got := EffectiveMessages(messages, 0, 0, 100000); len(got[2].Content) != len(big) {
		t.Error("pruned below the threshold")
	}
	got := EffectiveMessages(messages, 0, 0, 1200)
	if len(got[2].Content) >= len(big) || LargestToolResults(got, 1)[0].Bytes >= len(big) {
		t.Errorf("not pruned past the threshold: %d bytes", len(got[2].Content))
	}
	if len(messages[2].Content) != len(big) {
		t.Error("input changed")
	}
}

func TestEffectiveMessagesTrustsTheReportedSize(t *testing.T) {
	call := provider.ToolCall{ID: "1"}
	call.Function.Name = "read"
	big := strings.Repeat("x", 400)
	messages := []provider.Message{{Role: "user", Content: "first"}, {Role: "assistant", ToolCalls: []provider.ToolCall{call}}, {Role: "tool", ToolCallID: "1", Content: big}}
	for range pruneKeepTurns {
		messages = append(messages, provider.Message{Role: "user", Content: "next"})
	}
	if got := EffectiveMessages(messages, 0, 0, 1000); len(got[2].Content) != len(big) {
		t.Error("pruned although the estimate is far below the threshold")
	}
	if got := EffectiveMessages(messages, 0, 800, 1000); len(got[2].Content) >= len(big) {
		t.Error("the reported size past the threshold did not prune")
	}
}
