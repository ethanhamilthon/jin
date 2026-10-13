package session

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"jin/internal/core"
)

func TestTitleStartsAtFirstAcceptedPromptWithCapturedEffort(t *testing.T) {
	var titleCalls atomic.Int32
	started, release := make(chan struct{}), make(chan struct{})
	m, events, dir := newManagerWith(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "Write a title of 2 to 6 words") {
			titleCalls.Add(1)
			if !strings.Contains(string(body), "immutable first prompt") || !strings.Contains(string(body), "high") {
				t.Errorf("title request did not capture first prompt and effort: %s", body)
			}
			close(started)
			<-release
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"First topic\"}}]}\n\ndata: [DONE]\n\n")
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"answer\"}}]}\n\ndata: [DONE]\n\n")
	})
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	snap, _ := m.Create(dir)
	waitFor(t, events, func(e Event) bool { return e.Type == "state" && e.State.Ready })
	if err := m.SetModel(snap.State.ID, "m", "high"); err != nil {
		t.Fatal(err)
	}
	if err := m.Send(snap.State.ID, "immutable first prompt", nil, nil); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("title request did not start")
	}
	waitFor(t, events, func(e Event) bool { return e.Type == "ring" && e.Kind == core.UpdateDone })
	if titleCalls.Load() != 1 {
		t.Fatalf("title requests = %d", titleCalls.Load())
	}
	select {
	case <-release:
	default:
		close(release)
	}
}
