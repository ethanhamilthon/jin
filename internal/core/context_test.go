package core

import (
	"testing"

	"jin/internal/provider"
)

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
