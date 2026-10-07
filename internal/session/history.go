package session

import (
	"fmt"
	"strings"

	"jin/internal/core"
	"jin/internal/prompts"
	"jin/internal/provider"
	"jin/internal/tools"
)

// History rebuilds the entries a live session showed from its saved
// messages: user prompts, reasoning and replies, one entry per tool call
// with the summary the tool gave live, and the output of bash, edit and
// write calls.
func History(messages []provider.Message, registry *tools.Registry) []Entry {
	var entries []Entry
	calls := map[string]provider.ToolCall{}
	joinNext := false
	for _, msg := range messages {
		if msg.Role == "user" && core.IsTasksNote(msg.Content) {
			joinNext = true
			continue
		}
		if joinNext && msg.Role == "assistant" && msg.Content != "" && len(entries) > 0 && entries[len(entries)-1].Kind == core.UpdateAssistant {
			entries[len(entries)-1].Text += "\n\n" + msg.Content
			msg.Content, msg.ReasoningContent = "", ""
		}
		joinNext = false
		if msg.Role == "tool" {
			call := calls[msg.ToolCallID]
			if lines, ok := SavedResultLines(call, msg.Content); ok {
				entries = append(entries, Entry{Kind: core.UpdateToolResult, Tool: call.Function.Name, Lines: toLines(lines)})
			}
			if entry, ok := toolResultEntry(call, msg.Content); ok {
				entries = append(entries, entry)
			}
			continue
		}
		if msg.Role == "assistant" || msg.Role == "user" {
			entries = append(entries, messageEntries(msg, registry, calls)...)
		}
	}
	return entries
}

func messageEntries(msg provider.Message, registry *tools.Registry, calls map[string]provider.ToolCall) []Entry {
	switch {
	case core.IsSummary(msg):
		return []Entry{{Kind: core.UpdateCompacted, Text: CompactedLabel}}
	case core.ImageLabels(msg) != nil:
		var entries []Entry
		for i, label := range core.ImageLabels(msg) {
			entries = append(entries, pictureEntry(label, &msg.Images[i]))
		}
		return entries
	case msg.Role == "user" && strings.HasPrefix(msg.Content, "<task-result "):
		return []Entry{TaskEntry(msg.Content)}
	case msg.Role == "user":
		return []Entry{userEntry(userText(msg), msg.Images)}
	}
	var entries []Entry
	if msg.ReasoningContent != "" {
		entries = append(entries, Entry{Kind: core.UpdateReasoning, Text: msg.ReasoningContent})
	}
	if msg.Content != "" {
		entries = append(entries, Entry{Kind: core.UpdateAssistant, Text: msg.Content})
	}
	for _, call := range msg.ToolCalls {
		calls[call.ID] = call
		if entry, ok := TellEntry(call); ok {
			if entry.Kind != "" {
				entries = append(entries, entry)
			}
			continue
		}
		entries = append(entries, Entry{Kind: core.UpdateToolCall, Tool: call.Function.Name, Text: ToolSummary(registry, call)})
	}
	return entries
}

// userText is what the chat shows for a user message; its pictures were sent
// as separate parts, so they come back as numbered labels.
func userText(msg provider.Message) string {
	text := prompts.Strip(core.StripNotes(msg.Content))
	for i := range msg.Images {
		text = strings.TrimSpace(text + "\n" + fmt.Sprintf("[image %02d]", i+1))
	}
	return text
}
