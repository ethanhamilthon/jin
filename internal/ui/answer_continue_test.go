package ui

import (
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
)

func TestReplyAfterTasksNoteJoinsTheAnswer(t *testing.T) {
	s := &chatSession{width: 80}
	s.appendDelta(core.UpdateAssistant, "Done.")
	s.showUpdate(core.Update{Kind: core.UpdateHistory, Message: provider.Message{Role: "assistant", Content: "Done."}})
	s.showUpdate(core.Update{Kind: core.UpdateHistory, Message: provider.Message{Role: "user", Content: "<background-tasks>\\nStill running\\n</background-tasks>"}})
	s.appendDelta(core.UpdateReasoning, "thinking")
	s.appendDelta(core.UpdateAssistant, "The dev server keeps running.")
	if len(s.history) != 1 || s.history[0].text != "Done.\n\nThe dev server keeps running." {
		t.Fatalf("history = %+v", s.history)
	}
}

func TestHistoryJoinsReplyAfterTasksNote(t *testing.T) {
	entries := historyToEntries([]provider.Message{
		{Role: "user", Content: "go"},
		{Role: "assistant", Content: "Done."},
		{Role: "user", Content: "<background-tasks>\nStill running\n</background-tasks>"},
		{Role: "assistant", Content: "Stopped the server.", ReasoningContent: "hm"},
	}, nil)
	if len(entries) != 2 || entries[1].text != "Done.\n\nStopped the server." {
		t.Fatalf("entries = %+v", entries)
	}
}
