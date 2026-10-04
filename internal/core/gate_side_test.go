package core

import (
	"errors"
	"strings"
	"testing"

	"jin/internal/provider"
)

func TestRequestGateCoversSideRequests(t *testing.T) {
	stop := errors.New("budget stop")
	cases := []struct {
		name    string
		replies []string
		stopAt  int
	}{
		{"before compaction", []string{"summary"}, 1},
		{"after compaction", []string{"summary", "answer"}, 2},
		{"before side retry", []string{"tool", "summary"}, 2},
		{"after side retry", []string{"tool", "summary", "answer"}, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			agent, fake := newFakeAgent(t, c.replies...)
			var seen []provider.Usage
			agent.SetRequestGate(func(last provider.Usage) error {
				seen = append(seen, last)
				if len(seen) == c.stopAt {
					return stop
				}
				return nil
			})
			history := []provider.Message{{Role: "system", Content: "sys"}, {Role: "user", Content: "old"}}
			agent.SetContextSize(90)
			updates := make(chan Update, 256)
			err := agent.perform(t.Context(), t.Context(), Request{Model: "m", Window: 100, Prompt: "go"}, &history, nil, updates)
			if !errors.Is(err, stop) || fake.count() != c.stopAt-1 {
				t.Fatalf("err %v requests %d, want %d", err, fake.count(), c.stopAt-1)
			}
			if seen[0] != (provider.Usage{}) {
				t.Errorf("first gate usage = %+v", seen[0])
			}
			for _, usage := range seen[1:] {
				if !usage.Known || usage.Input != 10 {
					t.Errorf("side response usage not passed: %+v", usage)
				}
			}
			for _, u := range collect(updates) {
				if u.Kind == UpdateError && strings.Contains(u.Text, "Auto-compaction failed") {
					t.Errorf("refusal was treated as recoverable: %+v", u)
				}
			}
		})
	}
}
