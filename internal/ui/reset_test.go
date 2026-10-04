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
	s.showUpdate(core.Update{Kind: core.UpdateReset})
	s.showUpdate(core.Update{Kind: core.UpdateUsage, Model: "m", Usage: failed})
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

func streamed(s *chatSession, kind core.UpdateKind, text string) {
	s.showUpdate(core.Update{Kind: kind, Text: text})
}

func TestResetAfterInterjection(t *testing.T) {
	for _, laterDelta := range []bool{false, true} {
		s := &chatSession{width: 80}
		s.appendEntry(chatEntry{kind: core.UpdateUser, text: "question"})
		streamed(s, core.UpdateAssistantDelta, "ghost answer")
		s.closeOpenEntry()
		s.appendEntry(chatEntry{kind: core.UpdateUser, text: "queued note"})
		if laterDelta {
			streamed(s, core.UpdateAssistantDelta, "ghost tail")
		}
		s.showUpdate(core.Update{Kind: core.UpdateReset})
		streamed(s, core.UpdateAssistantDelta, "real answer")
		s.closeOpenEntry()
		screen := plain(s.rows)
		if strings.Contains(screen, "ghost") || !strings.Contains(screen, "queued note") || !strings.Contains(screen, "real answer") {
			t.Errorf("laterDelta=%v:\n%s", laterDelta, screen)
		}
	}
}

func TestResetAfterResizeAndFold(t *testing.T) {
	s := &chatSession{width: 80}
	s.appendEntry(chatEntry{kind: core.UpdateUser, text: "question"})
	streamed(s, core.UpdateAssistantDelta, "ghost answer")
	s.resize(30)
	s.setFold(foldMessages)
	s.setFold(foldOutput)
	s.showUpdate(core.Update{Kind: core.UpdateReset})
	screen := plain(s.rows)
	if strings.Contains(screen, "ghost") || !strings.Contains(screen, "question") {
		t.Fatalf("screen:\n%s", screen)
	}
}
