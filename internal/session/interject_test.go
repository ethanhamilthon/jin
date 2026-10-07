package session

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"jin/internal/core"
)

const toolCall = `{"choices":[{"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"c1","type":"function","function":{"name":"bash","arguments":"{\"command\":\"sleep 0.5\"}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":5,"completion_tokens":1}}`

// TestInterjectionReleasesTheSession sends a second message while the first
// turn runs a tool: the agent folds it into that turn, which ends with one
// UpdateDone, and the session must not stay busy.
func TestInterjectionReleasesTheSession(t *testing.T) {
	m, events, dir := newManagerWith(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		chunk := toolCall
		if strings.Contains(string(body), `"role":"tool"`) {
			chunk = answer
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: " + chunk + "\n\ndata: [DONE]\n\n"))
	})
	snap, _ := m.Create(dir)
	id := snap.State.ID
	waitFor(t, events, func(ev Event) bool { return ev.Type == "state" && ev.State.Ready })
	if err := m.Send(id, "first", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, events, func(ev Event) bool { return ev.Type == "entry" && ev.Entry.Kind == core.UpdateToolCall })
	if err := m.Send(id, "second", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, events, func(ev Event) bool { return ev.Type == "ring" && ev.Kind == core.UpdateDone })
	_ = m.Do(id, func(s *Session) error {
		if s.busy() || s.inflight != 0 {
			t.Fatalf("after the answer busy = %v, inflight = %d, pending = %d", s.busy(), s.inflight, len(s.pending))
		}
		return nil
	})
}
