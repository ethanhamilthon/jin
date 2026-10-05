package ui

import (
	"encoding/json"
	"strings"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/todo"
	"jin/internal/tools"
)

func toolSummary(registry *tools.Registry, call provider.ToolCall) string {
	if tool, ok := registry.Get(call.Function.Name); ok {
		if summary, ok := tool.Summary(call.Function.Arguments); ok {
			return summary
		}
	}
	return call.Function.Arguments
}

// toolResultEntry rebuilds the timeline entries that only exist after a tool
// finished: the answers of ask_user and a todo list with everything done.
func toolResultEntry(call provider.ToolCall, result string) (chatEntry, bool) {
	switch call.Function.Name {
	case "ask_user":
		if strings.Contains(result, " → ") {
			return chatEntry{kind: core.UpdateAsk, text: "Questions\n" + result}, true
		}
	case "todo":
		var args struct {
			Items []todo.Item `json:"items"`
		}
		if json.Unmarshal([]byte(call.Function.Arguments), &args) != nil || !strings.HasPrefix(result, "Todo list saved:") {
			return chatEntry{}, false
		}
		if items, err := todo.Validate(args.Items); err == nil && todo.AllDone(items) {
			return chatEntry{kind: core.UpdateTodo, text: todoEntryText(items)}, true
		}
	}
	return chatEntry{}, false
}
