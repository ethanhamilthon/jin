package headless

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/store"
)

func TestSaveFailureIsReported(t *testing.T) {
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) {})
	h.db.Close()
	r := &runState{db: h.db, id: "x", save: true, out: newWriter("json", &h.out, &h.errOut)}
	var o outcome
	var usage store.Usage
	r.apply(core.Update{Kind: core.UpdateHistory, Message: provider.Message{Role: "assistant", Content: "hi"}}, &o, &usage)
	r.apply(core.Update{Kind: core.UpdateDone, Final: true}, &o, &usage)
	if code := r.finish(o, result{Text: "hi"}); code != exitError {
		t.Fatalf("code %d", code)
	}
	if !strings.Contains(h.errOut.String(), "jin: could not save the session:") {
		t.Fatalf("stderr %q", h.errOut.String())
	}
	var rec map[string]any
	_ = json.Unmarshal([]byte(strings.TrimSpace(lastLine(h.out.String()))), &rec)
	if rec["is_error"] != false || rec["save_error"] == nil || rec["save_error"] == "" {
		t.Fatalf("record %v", rec)
	}
}

func lastLine(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	return lines[len(lines)-1]
}
