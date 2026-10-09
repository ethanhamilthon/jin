package session

import (
	"strings"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/tasks"
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
// finished: the answers of ask_user.
func toolResultEntry(call provider.ToolCall, result string) (Entry, bool) {
	switch call.Function.Name {
	case "ask_user":
		if strings.Contains(result, " → ") {
			return Entry{Kind: core.UpdateAsk, Text: "Questions\n" + result}, true
		}
	}
	return Entry{}, false
}

// TaskEntry shows a background task result by its summary.
func TaskEntry(text string) Entry {
	summary, ok := tasks.Summary(text)
	if !ok {
		summary = text
	}
	return Entry{Kind: core.UpdateInfo, Tool: TaskTool, Text: summary}
}
