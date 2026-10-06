package session

import (
	"encoding/json"
	"fmt"
	"strings"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/tasks"
	"jin/internal/todo"
	"jin/internal/tools"
)

// ToolSummary is the one-line description of a call that its tool gives,
// else its raw arguments.
func ToolSummary(registry *tools.Registry, call provider.ToolCall) string {
	if tool, ok := registry.Get(call.Function.Name); ok {
		if summary, ok := tool.Summary(call.Function.Arguments); ok {
			return summary
		}
	}
	return call.Function.Arguments
}

// toolResultEntry rebuilds the entries that only exist after a tool
// finished: the answers of ask_user and a todo list with everything done.
func toolResultEntry(call provider.ToolCall, result string) (Entry, bool) {
	switch call.Function.Name {
	case "ask_user":
		if strings.Contains(result, " → ") {
			return Entry{Kind: core.UpdateAsk, Text: "Questions\n" + result}, true
		}
	case "todo":
		var args struct {
			Items []todo.Item `json:"items"`
		}
		if json.Unmarshal([]byte(call.Function.Arguments), &args) != nil || !strings.HasPrefix(result, "Todo list saved:") {
			return Entry{}, false
		}
		if items, err := todo.Validate(args.Items); err == nil && todo.AllDone(items) {
			return Entry{Kind: core.UpdateTodo, Text: TodoText(items)}, true
		}
	}
	return Entry{}, false
}

// TellEntry is how a saved tell_user call shows again: a message as it was
// shown, a suggestion not at all (an empty entry).
func TellEntry(call provider.ToolCall) (Entry, bool) {
	if call.Function.Name != "tell_user" {
		return Entry{}, false
	}
	var args struct{ Mode, Text string }
	_ = json.Unmarshal([]byte(call.Function.Arguments), &args)
	if args.Mode != "message" || strings.TrimSpace(args.Text) == "" {
		return Entry{}, true
	}
	return Entry{Kind: core.UpdateTell, Text: strings.TrimSpace(args.Text)}, true
}

// TodoText is the entry of a finished todo list.
func TodoText(items []todo.Item) string {
	done, total := todo.Counts(items)
	return fmt.Sprintf("Todo %d/%d\n%s", done, total, todo.Text(items))
}

// TaskEntry shows a background task result by its summary.
func TaskEntry(text string) Entry {
	summary, ok := tasks.Summary(text)
	if !ok {
		summary = text
	}
	return Entry{Kind: core.UpdateInfo, Tool: TaskTool, Text: summary}
}
