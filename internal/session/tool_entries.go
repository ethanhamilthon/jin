package session

import (
	"strings"

	"jin/internal/core"
	"jin/internal/diff"
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

// TaskEntries shows a background task result like a tool call: a call line
// with the task and its status, then the output.
func TaskEntries(text string) []Entry {
	headline, lines, ok := tasks.Parse(text)
	if !ok {
		return []Entry{{Kind: core.UpdateInfo, Text: text}}
	}
	out := make([]diff.Line, len(lines))
	for i, line := range lines {
		out[i] = diff.Line{Op: diff.Keep, Text: line}
	}
	return []Entry{
		{Kind: core.UpdateToolCall, Tool: TaskTool, Text: headline},
		{Kind: core.UpdateToolResult, Tool: TaskTool, Lines: toLines(out)},
	}
}
