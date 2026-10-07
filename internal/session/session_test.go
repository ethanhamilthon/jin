package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"jin/internal/core"
	"jin/internal/store"
)

const answer = `{"choices":[{"delta":{"role":"assistant","content":"hello there"}}],"usage":{"prompt_tokens":10,"completion_tokens":3}}`

func newTestManager(t *testing.T) (*Manager, chan Event, string) {
	return newManagerWith(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: " + answer + "\n\ndata: [DONE]\n\n"))
	})
}

func newManagerWith(t *testing.T, handler http.HandlerFunc) (*Manager, chan Event, string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	entry := store.ProviderEntry{ID: "p", Name: "p", Kind: "openai", BaseURL: server.URL, APIKey: "k"}
	if err := db.AddProvider(entry); err != nil {
		t.Fatal(err)
	}
	if err := db.SetActiveProvider("p"); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveModel("m", ""); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan Event, 1024)
	m := NewManager(ctx, db, "v0", func(ev Event) { events <- ev })
	t.Cleanup(func() { cancel(); m.Shutdown() })
	return m, events, t.TempDir()
}

func waitFor(t *testing.T, events chan Event, match func(Event) bool) {
	t.Helper()
	timeout := time.After(10 * time.Second)
	for {
		select {
		case ev := <-events:
			if match(ev) {
				return
			}
		case <-timeout:
			t.Fatal("timed out waiting for an event")
		}
	}
}

func TestSendStreamsAndSaves(t *testing.T) {
	m, events, dir := newTestManager(t)
	snap, err := m.Create(dir)
	if err != nil || snap.Intro == nil {
		t.Fatalf("create: %v %+v", err, snap)
	}
	id := snap.State.ID
	waitFor(t, events, func(ev Event) bool { return ev.Type == "state" && ev.State.Ready })
	if err := m.Send(id, "say hi", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, events, func(ev Event) bool { return ev.Type == "ring" && ev.Kind == core.UpdateDone })
	snap, _ = m.Snapshot(id)
	if len(snap.Entries) != 2 || snap.Entries[1].Text != "hello there" || !snap.State.Persisted || snap.State.Title != "say hi" {
		t.Fatalf("snapshot = %+v", snap)
	}
	if !m.Close(id) {
		t.Fatal("idle session did not close")
	}
	snap, err = m.Open(id)
	if err != nil || len(snap.Entries) != 2 || snap.Entries[0].Kind != core.UpdateUser || snap.State.Usage.Input != 10 {
		t.Fatalf("reopened = %v %+v", err, snap)
	}
}

func TestShellOutput(t *testing.T) {
	m, events, dir := newTestManager(t)
	snap, _ := m.Create(dir)
	if err := m.Shell(snap.State.ID, "echo hi"); err != nil {
		t.Fatal(err)
	}
	waitFor(t, events, func(ev Event) bool { return ev.Type == "entry" && ev.Entry.Text == "hi" })
}
