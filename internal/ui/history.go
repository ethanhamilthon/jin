package ui

import (
	"encoding/json"
	"strings"

	"jin/internal/core"
	"jin/internal/prompts"
	"jin/internal/provider"
	"jin/internal/todo"
	"jin/internal/tools"
)

// historyToEntries rebuilds the visible chat log from a persisted message
// trail, mirroring exactly what the live session showed: user prompts,
// reasoning followed by assistant replies, and one row per tool call using
// the same human-readable summary the tool produced live. Tool result
// messages are never shown, matching the live chat.
func historyToEntries(messages []provider.Message, registry *tools.Registry) []chatEntry {
	var entries []chatEntry
	calls := map[string]provider.ToolCall{}
	for _, msg := range messages {
		if msg.Role == "tool" {
			if entry, ok := toolResultEntry(calls[msg.ToolCallID], msg.Content); ok {
				entries = append(entries, entry)
			}
			continue
		}
		if msg.Role != "assistant" && msg.Role != "user" {
			continue
		}
		if core.IsSummary(msg) {
			entries = append(entries, chatEntry{kind: core.UpdateCompacted, text: compactedLabel})
			continue
		}
		if labels := core.ImageLabels(msg); labels != nil {
			for _, label := range labels {
				entries = append(entries, chatEntry{kind: core.UpdateInfo, text: label})
			}
			continue
		}
		if msg.Role == "user" {
			entries = append(entries, chatEntry{kind: core.UpdateUser, text: prompts.Strip(core.StripTodoEdited(msg.Content))})
			continue
		}
		if msg.ReasoningContent != "" {
			entries = append(entries, chatEntry{kind: core.UpdateReasoning, text: msg.ReasoningContent})
		}
		if msg.Content != "" {
			entries = append(entries, chatEntry{kind: core.UpdateAssistant, text: msg.Content})
		}
		for _, call := range msg.ToolCalls {
			calls[call.ID] = call
			entries = append(entries, chatEntry{kind: core.UpdateToolCall, tool: call.Function.Name, text: toolSummary(registry, call)})
		}
	}
	return entries
}

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
