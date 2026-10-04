package ui

import (
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/store"
)

func TestResetDropsFailedAttemptAndCountsItsUsage(t *testing.T) {
	s := &chatSession{width: 80}
	s.appendEntry(chatEntry{kind: core.UpdateUser, text: "question"})
	s.showUpdate(core.Update{Kind: core.UpdateReasoningDelta, Text: "ghost thoughts"})
	s.showUpdate(core.Update{Kind: core.UpdateAssistantDelta, Text: "ghost answer"})
	failed := provider.Usage{Input: 100, Output: 5, Known: true}
	s.showUpdate(core.Update{Kind: core.UpdateReset, Model: "m", Usage: failed})
	s.showUpdate(core.Update{Kind: core.UpdateInfo, Text: "retrying"})
	s.showUpdate(core.Update{Kind: core.UpdateAssistantDelta, Text: "real answer"})
	s.closeOpenEntry()
	screen := plain(s.rows)
	if strings.Contains(screen, "ghost") || !strings.Contains(screen, "real answer") {
		t.Fatalf("screen after retry:\n%s", screen)
	}
	if len(s.history) != 3 {
		t.Fatalf("history = %+v", s.history)
	}
	if s.usage != (store.Usage{Input: 100, Output: 5, Context: s.usage.Context, Cost: s.usage.Cost}) {
		t.Errorf("usage = %+v", s.usage)
	}
}

func TestResetWithoutDeltasChangesNothing(t *testing.T) {
	s := &chatSession{width: 80}
	s.appendEntry(chatEntry{kind: core.UpdateUser, text: "question"})
	s.showUpdate(core.Update{Kind: core.UpdateReset})
	if len(s.history) != 1 {
		t.Fatalf("history = %+v", s.history)
	}
}
