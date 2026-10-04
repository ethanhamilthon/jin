package headless

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/store"
)

func TestHeadlessWriteCanBeUndone(t *testing.T) {
	file := filepath.Join(t.TempDir(), "out.txt")
	calls := 0
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			args, _ := json.Marshal(map[string]string{"path": file, "content": "new"})
			call, _ := json.Marshal(string(args))
			sse(w, fmt.Sprintf(`{"choices":[{"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"c1","type":"function","function":{"name":"write","arguments":%s}}]}}]}`, call))
			return
		}
		sse(w, answerChunk)
	})
	if code := h.run(t, "-p", "write it"); code != 0 {
		t.Fatalf("code %d: %s", code, h.errOut.String())
	}
	list, _ := h.db.ListByPath(h.dir)
	turn, changes, err := h.db.LastChanges(list[0].ID)
	if err != nil || turn != 1 || len(changes) != 1 || changes[0].After != "new" {
		t.Fatalf("turn %d changes %+v err %v", turn, changes, err)
	}
}

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
