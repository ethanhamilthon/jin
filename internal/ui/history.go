package ui

import (
	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/session"
	"jin/internal/tools"
)

// historyToEntries rebuilds the visible chat log from a persisted message
// trail, mirroring exactly what the live session showed.
func historyToEntries(messages []provider.Message, registry *tools.Registry) []chatEntry {
	var entries []chatEntry
	for _, entry := range session.History(messages, registry) {
		entries = append(entries, chatEntryOf(entry))
	}
	return entries
}

func chatEntryOf(entry session.Entry) chatEntry {
	if entry.Kind == core.UpdateToolResult {
		return chatEntry{kind: entry.Kind, tool: entry.Tool, text: tailText(session.DiffLines(entry.Lines))}
	}
	return chatEntry{kind: entry.Kind, tool: entry.Tool, text: entry.Text}
}

func toolSummary(registry *tools.Registry, call provider.ToolCall) string {
	return session.ToolSummary(registry, call)
}
