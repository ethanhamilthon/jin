package ui

import (
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/todo"
	"jin/internal/tools"
)

func call(id, name, args string) provider.Message {
	c := provider.ToolCall{ID: id}
	c.Function.Name, c.Function.Arguments = name, args
	return provider.Message{Role: "assistant", ToolCalls: []provider.ToolCall{c}}
}

func TestHistoryRebuildsAskAndFinishedTodo(t *testing.T) {
	msgs := []provider.Message{
		call("1", "ask_user", `{"questions":[{"question":"Ok?"}]}`),
		{Role: "tool", ToolCallID: "1", Content: "Ok? → yes"},
		call("2", "todo", `{"items":[{"text":"a","status":"done"}]}`),
		{Role: "tool", ToolCallID: "2", Content: "Todo list saved:\n- [x] a"},
		call("3", "todo", `{"items":[{"text":"a","status":"pending"}]}`),
		{Role: "tool", ToolCallID: "3", Content: "Todo list saved:\n- [ ] a"},
		{Role: "user", Content: core.TodoEditedBlock([]todo.Item{{Text: "a", Status: todo.Done}}) + "hi"},
	}
	var asks, todos int
	var last chatEntry
	for _, e := range historyToEntries(msgs, tools.Build(tools.Catalog(), &tools.MemoryTodos{})) {
		switch e.kind {
		case core.UpdateAsk:
			asks++
		case core.UpdateTodo:
			todos++
		}
		last = e
	}
	if asks != 1 || todos != 1 {
		t.Fatalf("asks=%d todos=%d", asks, todos)
	}
	if last.text != "hi" {
		t.Fatalf("todo-edited block shown: %q", last.text)
	}
}
