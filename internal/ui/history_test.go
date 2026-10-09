package ui

import (
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/tools"
)

func call(id, name, args string) provider.Message {
	c := provider.ToolCall{ID: id}
	c.Function.Name, c.Function.Arguments = name, args
	return provider.Message{Role: "assistant", ToolCalls: []provider.ToolCall{c}}
}

func TestHistoryRebuildsAsk(t *testing.T) {
	msgs := []provider.Message{
		call("1", "ask_user", `{"questions":[{"question":"Ok?"}]}`),
		{Role: "tool", ToolCallID: "1", Content: "Ok? → yes"},
		{Role: "user", Content: "<todo-edited>The user edited the todo list.</todo-edited>\n\nhi"},
	}
	var asks int
	var last chatEntry
	for _, e := range historyToEntries(msgs, tools.Build(tools.Catalog())) {
		if e.kind == core.UpdateAsk {
			asks++
		}
		last = e
	}
	if asks != 1 {
		t.Fatalf("asks=%d", asks)
	}
	if last.text != "hi" {
		t.Fatalf("todo-edited block shown: %q", last.text)
	}
}

func TestHistoryShowsRemovedToolsAsPlainCalls(t *testing.T) {
	msgs := []provider.Message{
		call("1", "todo", `{"items":[{"text":"a"}]}`),
		{Role: "tool", ToolCallID: "1", Content: "Todo list saved:\n- [ ] a"},
		call("2", "tell_user", `{"mode":"message","text":"hi"}`),
		{Role: "tool", ToolCallID: "2", Content: "ok"},
	}
	var calls int
	for _, e := range historyToEntries(msgs, tools.Build(tools.Catalog())) {
		if e.kind == core.UpdateToolCall {
			calls++
		}
	}
	if calls != 2 {
		t.Fatalf("calls shown = %d, want 2", calls)
	}
}
