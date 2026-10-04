package ui

import (
	"strings"
	"testing"

	"jin/internal/core"
)

func TestStreamRendersOncePerFrame(t *testing.T) {
	s := &chatSession{width: 80}
	const deltas = 3000
	for range deltas {
		s.showUpdate(core.Update{Kind: core.UpdateAssistantDelta, Text: "word **bold** and `code` "})
	}
	if s.stream.renders > deltas/20 {
		t.Fatalf("rendered %d times for %d deltas", s.stream.renders, deltas)
	}
	s.closeOpenEntry()
	want := strings.TrimSpace(strings.Repeat("word **bold** and `code` ", deltas))
	if got := s.history[0].text; got != want {
		t.Fatalf("final text differs: %d vs %d bytes", len(got), len(want))
	}
	if len(s.rows) != len(entryRows(s.history[0], 80)) {
		t.Fatal("final rows do not match a full render")
	}
}
