package headless

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"

	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
)

// Compaction and the following answer are both gated model requests.
func TestCompactionCountsAsATurn(t *testing.T) {
	cases := []struct {
		maxTurns string
		code     int
	}{{"3", 0}, {"2", exitBudget}}
	for _, c := range cases {
		var hits atomic.Int32
		h := newHarness(t, loopingServer(&hits, "true", 1))
		h.dir = t.TempDir()
		loadPricing = func(context.Context) pricing.Table {
			return pricing.Table{"m": {InputCostPerToken: 1, OutputCostPerToken: 2, MaxInputTokens: 100}}
		}
		if err := h.db.TouchProvider("s1", h.dir, "m", "", "t", ""); err != nil {
			t.Fatal(err)
		}
		for _, msg := range []provider.Message{{Role: "user", Content: "old"}, {Role: "assistant", Content: "older"}} {
			if err := h.db.AppendMessage("s1", msg); err != nil {
				t.Fatal(err)
			}
		}
		if err := h.db.SaveUsage("s1", store.Usage{Context: 90}); err != nil {
			t.Fatal(err)
		}
		code := h.run(t, "-p", "--session", "s1", "--max-turns", c.maxTurns, "go")
		if code != c.code || !strings.Contains(h.errOut.String(), "compact") {
			t.Errorf("--max-turns %s: code %d, want %d; stderr %q", c.maxTurns, code, c.code, h.errOut.String())
		}
	}
}
