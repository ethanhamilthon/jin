package headless

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
)

func TestMaxCostCountsBilledRetryUsage(t *testing.T) {
	cases := []struct {
		name, limit, text string
		compaction        bool
		requests, code    int
	}{
		{"failed answer plus tool uses budget", "20", "working on it", false, 2, exitBudget},
		{"successful response is not counted twice", "49", "hello there", false, 3, exitOK},
		{"failed compaction plus summary uses budget", "20", "", true, 2, exitBudget},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var hits atomic.Int32
			h := newHarness(t, func(w http.ResponseWriter, r *http.Request) {
				n := hits.Add(1)
				if n == 1 {
					w.Header().Set("Content-Length", "100000")
					fmt.Fprintf(w, "data: %s\n\n", answerChunk)
					return
				}
				if !c.compaction && n == 2 {
					sse(w, toolCallChunk("true"))
					return
				}
				sse(w, answerChunk)
			})
			h.dir = t.TempDir()
			args := []string{"-p", "--format", "json", "--max-cost", c.limit}
			if c.compaction {
				loadPricing = func(context.Context) pricing.Table {
					return pricing.Table{"m": {InputCostPerToken: 1, OutputCostPerToken: 2, MaxInputTokens: 100000}}
				}
				if err := h.db.TouchProvider("s1", h.dir, "m", "", "t", ""); err != nil {
					t.Fatal(err)
				}
				if err := h.db.AppendMessage("s1", provider.Message{Role: "user", Content: "old"}); err != nil {
					t.Fatal(err)
				}
				if err := h.db.SaveUsage("s1", store.Usage{Context: 90000}); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--session", "s1")
			}
			if code := h.run(t, append(args, "go")...); code != c.code || int(hits.Load()) != c.requests {
				t.Fatalf("code %d requests %d, want %d/%d; stderr %q", code, hits.Load(), c.code, c.requests, h.errOut.String())
			}
			rec := lastRecord(t, h.out.String())
			if rec["result"] != c.text || rec["is_error"] != (c.code == exitBudget) {
				t.Errorf("result = %v", rec)
			}
			if c.code == exitBudget && rec["error"] != "budget reached: max-cost" {
				t.Errorf("error = %v", rec["error"])
			}
			session, _, err := h.db.GetSession(rec["session_id"].(string))
			if err != nil || session.Usage.Cost != float64(16*c.requests) ||
				rec["usage"].(map[string]any)["cost"] != float64(16*c.requests) {
				t.Errorf("session %+v error %v JSON %v", session, err, rec)
			}
			if !c.compaction && c.code == exitBudget {
				messages, err := h.db.LoadMessages(session.ID)
				if err != nil || messages[len(messages)-1].Role != "tool" {
					t.Errorf("unclosed tool call: messages %+v err %v", messages, err)
				}
			}
			if !strings.Contains(h.errOut.String(), "retrying") {
				t.Errorf("retry not exercised: %s", h.errOut.String())
			}
		})
	}
}
