package core

import (
	"errors"
	"testing"

	"jin/internal/provider"
)

func TestRequestGateRunsBeforeEveryRequest(t *testing.T) {
	stop := errors.New("stop")
	cases := []struct {
		name     string
		stopAt   int
		requests int
		final    bool
	}{
		{"stops before the first request", 1, 0, false},
		{"stops after a tool result", 2, 1, false},
		{"never stops", 0, 3, true},
	}
	for _, c := range cases {
		agent, fake := newFakeAgent(t, "tool", "tool", "answer")
		var seen []provider.Usage
		agent.SetRequestGate(func(last provider.Usage) error {
			seen = append(seen, last)
			if len(seen) == c.stopAt {
				return stop
			}
			return nil
		})
		history := []provider.Message{{Role: "system", Content: "sys"}}
		updates := make(chan Update, 256)
		agent.turn(t.Context(), Request{Prompt: "go", Model: "m"}, &history, nil, updates)
		var done Update
		for _, u := range collect(updates) {
			if u.Kind == UpdateDone {
				done = u
			}
		}
		if fake.count() != c.requests || done.Final != c.final {
			t.Errorf("%s: requests = %d, final = %v", c.name, fake.count(), done.Final)
		}
		if seen[0] != (provider.Usage{}) {
			t.Errorf("%s: first gate call got %+v, want zero usage", c.name, seen[0])
		}
		if len(seen) > 1 && (seen[1].Input != 10 || seen[1].Output != 2) {
			t.Errorf("%s: second gate call got %+v", c.name, seen[1])
		}
		if last := history[len(history)-1]; c.stopAt == 2 && last.Role != "tool" {
			t.Errorf("%s: history ends with %+v, want the tool result", c.name, last)
		}
	}
}
