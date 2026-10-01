package ui

import (
	"jin/internal/core"
	"jin/internal/prompts"
	"jin/internal/provider"
	"jin/internal/tools"
)

// historyToEntries rebuilds the visible chat log from a persisted message
// trail, mirroring exactly what the live session showed: user prompts,
// reasoning followed by assistant replies, and one row per tool call using
// the same human-readable summary the tool produced live. Tool result
// messages are never shown, matching the live chat.
func historyToEntries(messages []provider.Message, registry *tools.Registry) []chatEntry {
	var entries []chatEntry
	for _, msg := range messages {
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
			entries = append(entries, chatEntry{kind: core.UpdateUser, text: prompts.Strip(msg.Content)})
			continue
		}
		if msg.ReasoningContent != "" {
			entries = append(entries, chatEntry{kind: core.UpdateReasoning, text: msg.ReasoningContent})
		}
		if msg.Content != "" {
			entries = append(entries, chatEntry{kind: core.UpdateAssistant, text: msg.Content})
		}
		for _, call := range msg.ToolCalls {
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
