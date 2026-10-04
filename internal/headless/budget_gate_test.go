package headless

import (
	"testing"

	"jin/internal/core"
	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
)

func TestGateAllowsExactlyTheBudget(t *testing.T) {
	used := provider.Usage{Known: true, Input: 10, Output: 3}
	cases := []struct {
		name    string
		opt     Options
		allowed int
		reason  string
	}{
		{"no limit", Options{}, 5, ""},
		{"max-turns 2", Options{MaxTurns: 2}, 2, "max-turns"},
		{"max-cost 20 (16 per request)", Options{MaxCost: 20}, 2, "max-cost"},
		{"max-cost 16 stops at the limit", Options{MaxCost: 16}, 1, "max-cost"},
	}
	for _, c := range cases {
		r := &runState{budget: newBudget(c.opt, store.Usage{Cost: 100}), request: core.Request{Model: "m"},
			table: pricing.Table{"m": {InputCostPerToken: 1, OutputCostPerToken: 2}}}
		allowed, last := 0, provider.Usage{}
		for range 5 {
			if r.gate(last) != nil {
				break
			}
			allowed, last = allowed+1, used
		}
		if allowed != c.allowed || r.budget.result(store.Usage{}) != c.reason || r.budget.halted() != (c.reason != "") {
			t.Errorf("%s: allowed %d, result %q", c.name, allowed, r.budget.result(store.Usage{}))
		}
	}
}
