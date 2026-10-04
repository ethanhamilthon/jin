package core

import (
	"strings"
	"testing"

	"jin/internal/provider"
)

func TestPruneToolResults(t *testing.T) {
	history := toolTurns(3, 400)
	history = append(history, imageMessage(nil))
	small := toolTurns(3, 10)
	done := toolTurns(3, 400)
	done[3].Content = omittedNote("read", 999)
	cases := []struct {
		name    string
		history []provider.Message
		keep    int
		omitted int
	}{
		{"keeps the last turns", history, 2, 1},
		{"image messages are no turns", history, 1, 2},
		{"too few turns", history, 3, 0},
		{"small results stay", small, 1, 0},
		{"omitted stays as is", done, 1, 2},
	}
	for _, c := range cases {
		pruned, changed := pruneToolResults(c.history, c.keep)
		if omitted(pruned) != c.omitted || changed != (omitted(pruned) > omitted(c.history)) {
			t.Errorf("%s: omitted = %d, changed = %v, want %d", c.name, omitted(pruned), changed, c.omitted)
		}
		if len(pruned) != len(c.history) {
			t.Fatalf("%s: length changed", c.name)
		}
		for i := range pruned {
			if pruned[i].Role != c.history[i].Role || pruned[i].ToolCallID != c.history[i].ToolCallID || len(pruned[i].ToolCalls) != len(c.history[i].ToolCalls) {
				t.Errorf("%s: message %d lost its pairing", c.name, i)
			}
		}
	}
	if pruned, _ := pruneToolResults(done, 1); pruned[3].Content != done[3].Content {
		t.Errorf("an omitted result was rewritten: %q", pruned[3].Content)
	}
	if !strings.HasPrefix(history[3].Content, "xxx") {
		t.Error("pruning changed the input history")
	}
	if note := omittedNote("read", 400); note != "[output of read omitted, 400 bytes; run it again if needed]" {
		t.Errorf("note = %q", note)
	}
}

func TestPruneLowersTheEstimate(t *testing.T) {
	agent, _ := newFakeAgent(t, "unused")
	history := toolTurns(3, 4000)
	agent.reported(provider.Usage{Known: true, Input: 5000}, history)
	if !agent.prune(&history, 1) || agent.size != 5000-2*(4000-len(omittedNote("read", 4000)))/4 {
		t.Errorf("size = %d after prune", agent.size)
	}
}
