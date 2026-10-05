package ui

import (
	"strings"

	"jin/internal/core"
	"jin/internal/prompts"
	"jin/internal/provider"
	"jin/internal/tools"
)

// historyToEntries rebuilds the visible chat log from a persisted message
// trail, mirroring exactly what the live session showed: user prompts,
// reasoning followed by assistant replies, and one row per tool call using
// the same human-readable summary the tool produced live. bash, edit and
// write results come back as tool output entries.
func historyToEntries(messages []provider.Message, registry *tools.Registry) []chatEntry {
	var entries []chatEntry
	calls := map[string]provider.ToolCall{}
	joinNext := false
	for _, msg := range messages {
		if msg.Role == "user" && core.IsTasksNote(msg.Content) {
			joinNext = true
			continue
		}
		if joinNext && msg.Role == "assistant" && msg.Content != "" && len(entries) > 0 && entries[len(entries)-1].kind == core.UpdateAssistant {
			entries[len(entries)-1].text += "\n\n" + msg.Content
			msg.Content, msg.ReasoningContent = "", ""
		}
		joinNext = false
		if msg.Role == "tool" {
			if text, ok := savedResultText(calls[msg.ToolCallID], msg.Content); ok {
				entries = append(entries, chatEntry{kind: core.UpdateToolResult, tool: calls[msg.ToolCallID].Function.Name, text: text})
			}
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
		if msg.Role == "user" && strings.HasPrefix(msg.Content, "<task-result ") {
			entries = append(entries, taskChatEntry(msg.Content))
			continue
		}
		if msg.Role == "user" {
			entries = append(entries, chatEntry{kind: core.UpdateUser, text: prompts.Strip(core.StripNotes(msg.Content))})
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
			if entry, ok := tellEntry(call); ok {
				if entry.kind != "" {
					entries = append(entries, entry)
				}
				continue
			}
			entries = append(entries, chatEntry{kind: core.UpdateToolCall, tool: call.Function.Name, text: toolSummary(registry, call)})
		}
	}
	return entries
}
