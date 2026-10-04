package headless

import (
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
)

func TestBudgetGatesAutomaticCompaction(t *testing.T) {
	cases := []struct {
		name, flag, limit string
		initial           bool
		requests, code    int
	}{
		{"refuse compaction after last turn", "max-turns", "1", false, 1, exitBudget},
		{"refuse compaction after spent cost", "max-cost", "90006", false, 1, exitBudget},
		{"compaction consumes last turn", "max-turns", "2", false, 2, exitBudget},
		{"compaction consumes remaining cost", "max-cost", "90010", false, 2, exitBudget},
		{"initial compaction consumes last turn", "max-turns", "1", true, 1, exitBudget},
		{"initial compaction consumes remaining cost", "max-cost", "16", true, 1, exitBudget},
		{"cost is not counted twice", "max-cost", "90040", false, 3, exitOK},
		{"initial cost is not counted twice", "max-cost", "33", true, 2, exitOK},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var hits atomic.Int32
			h := newHarness(t, func(w http.ResponseWriter, r *http.Request) {
				if n := hits.Add(1); !c.initial && n == 1 {
					sse(w, strings.Replace(toolCallChunk("true"), `"prompt_tokens":10`, `"prompt_tokens":90000`, 1))
					return
				}
				sse(w, answerChunk)
			})
			h.dir = t.TempDir()
			loadPricing = func(context.Context) pricing.Table {
				return pricing.Table{"m": {InputCostPerToken: 1, OutputCostPerToken: 2, MaxInputTokens: 100000}}
			}
			args := []string{"-p", "--format", "json", "--" + c.flag, c.limit}
			if c.initial {
				if err := h.db.TouchProvider("s1", h.dir, "m", "", "t", ""); err != nil {
					t.Fatal(err)
				}
				for _, msg := range []provider.Message{{Role: "user", Content: "old"}, {Role: "assistant", Content: "older"}} {
					if err := h.db.AppendMessage("s1", msg); err != nil {
						t.Fatal(err)
					}
				}
				if err := h.db.SaveUsage("s1", store.Usage{Context: 90000}); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--session", "s1")
			}
			if code := h.run(t, append(args, "go")...); code != c.code || int(hits.Load()) != c.requests {
				t.Fatalf("code %d requests %d, want %d/%d; stderr %q", code, hits.Load(), c.code, c.requests, h.errOut.String())
			}
			if strings.Contains(h.errOut.String(), "Auto-compaction failed:") {
				t.Fatalf("budget refusal treated as compaction failure: %s", h.errOut.String())
			}
			rec := lastRecord(t, h.out.String())
			if c.code == exitBudget && (rec["error"] != "budget reached: "+c.flag || rec["is_error"] != true) {
				t.Errorf("result = %v", rec)
			}
			if !c.initial && c.requests == 1 {
				messages, err := h.db.LoadMessages(rec["session_id"].(string))
				if err != nil || messages[len(messages)-1].Role != "tool" {
					t.Errorf("unclosed tool call: messages %+v err %v", messages, err)
				}
			}
		})
	}
}
