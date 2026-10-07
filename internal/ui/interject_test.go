package ui

import (
	"testing"

	"jin/internal/core"
)

// TestInterjectionReleasesTheSession hands two requests to the agent; the
// second is folded into the first turn, so the turn ends with UpdateTaken
// and one UpdateDone.
func TestInterjectionReleasesTheSession(t *testing.T) {
	a, s := ownedTestSession(t)
	s.prompts = make(chan core.Request, 2)
	s.pending = []core.Request{{Prompt: "first"}, {Prompt: "second"}}
	a.flushPending()
	if s.inflight != 2 {
		t.Fatalf("inflight = %d, want 2", s.inflight)
	}
	entries := len(s.history)
	a.applyUpdate(s.id, core.Update{Kind: core.UpdateWorking})
	a.applyUpdate(s.id, core.Update{Kind: core.UpdateTaken})
	a.applyUpdate(s.id, core.Update{Kind: core.UpdateDone, Final: true})
	if s.inflight != 0 || a.anyWorking() || len(s.history) != entries {
		t.Fatalf("inflight = %d, working = %v, entries %d → %d", s.inflight, a.anyWorking(), entries, len(s.history))
	}
	assertOwned(t, a, s, false)
}
