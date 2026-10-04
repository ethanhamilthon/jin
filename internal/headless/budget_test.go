package headless

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

// toolCallChunk asks for a bash command and reports usage of 10 input and
// 3 output tokens: 16 dollars at the prices of the test catalogue.
func toolCallChunk(command string) string {
	args, _ := json.Marshal(map[string]string{"command": command})
	quoted, _ := json.Marshal(string(args))
	return `{"choices":[{"delta":{"role":"assistant","content":"working on it","tool_calls":[{"index":0,"id":"c1","type":"function","function":{"name":"bash","arguments":` +
		string(quoted) + `}}]}}],"usage":{"prompt_tokens":10,"completion_tokens":3}}`
}

// loopingServer answers every request with a tool call, and with the final
// answer from request number finalAt on (0: never).
func loopingServer(hits *atomic.Int32, command string, finalAt int32) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if finalAt > 0 && n >= finalAt {
			sse(w, answerChunk)
			return
		}
		sse(w, toolCallChunk(command))
	}
}

func lastRecord(t *testing.T, out string) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var rec map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &rec); err != nil {
		t.Fatalf("last line %q: %v", lines[len(lines)-1], err)
	}
	return rec
}

func TestMaxTurnsStopsTheRunWithExitThree(t *testing.T) {
	var hits atomic.Int32
	h := newHarness(t, loopingServer(&hits, "true", 0))
	h.dir = t.TempDir()
	if code := h.run(t, "-p", "--max-turns", "2", "go"); code != exitBudget {
		t.Fatalf("code %d, stderr %q", code, h.errOut.String())
	}
	if hits.Load() != 2 {
		t.Errorf("model requests = %d, want 2", hits.Load())
	}
	if h.out.String() != "working on it\n" || !strings.Contains(h.errOut.String(), "budget reached: max-turns") {
		t.Errorf("stdout %q stderr %q", h.out.String(), h.errOut.String())
	}
	list, _ := h.db.ListByPath(h.dir)
	if len(list) != 1 || list[0].Usage.Input != 20 {
		t.Errorf("session not saved with its usage: %+v", list)
	}
	msgs, _ := h.db.LoadMessages(list[0].ID)
	if last := msgs[len(msgs)-1]; last.Role != "tool" {
		t.Errorf("history must end with a closed tool call, got %+v", last)
	}
}

func TestMaxCostStopsTheRunAndJSONCarriesTheError(t *testing.T) {
	var hits atomic.Int32
	h := newHarness(t, loopingServer(&hits, "true", 0))
	h.dir = t.TempDir()
	if code := h.run(t, "-p", "--format", "json", "--max-cost", "20", "go"); code != exitBudget {
		t.Fatalf("code %d, stderr %q", code, h.errOut.String())
	}
	rec := lastRecord(t, h.out.String())
	if rec["error"] != "budget reached: max-cost" || rec["is_error"] != true || rec["result"] != "working on it" {
		t.Errorf("result = %v", rec)
	}
	if hits.Load() != 2 {
		t.Errorf("model requests = %d, want 2 (cost 16 then 32)", hits.Load())
	}
}

func TestBudgetReachedByTheFinalAnswerStillGivesExitThree(t *testing.T) {
	var hits atomic.Int32
	h := newHarness(t, loopingServer(&hits, "true", 2))
	h.dir = t.TempDir()
	if code := h.run(t, "-p", "--format", "json", "--max-turns", "2", "go"); code != exitBudget {
		t.Fatalf("code %d stdout %q stderr %q", code, h.out.String(), h.errOut.String())
	}
	if rec := lastRecord(t, h.out.String()); rec["error"] != "budget reached: max-turns" || rec["is_error"] != true || rec["result"] != "hello there" {
		t.Errorf("result = %v", rec)
	}
}

func TestBudgetCountsOnlyThisRun(t *testing.T) {
	var hits atomic.Int32
	h := newHarness(t, loopingServer(&hits, "true", 1))
	h.dir = t.TempDir()
	for i := range 2 {
		args := []string{"-p", "--max-cost", "20", "go"}
		if i > 0 {
			args = append(args[:1], append([]string{"-c"}, args[1:]...)...)
		}
		if code := h.run(t, args...); code != 0 {
			t.Fatalf("run %d: code %d stderr %q", i, code, h.errOut.String())
		}
	}
}
