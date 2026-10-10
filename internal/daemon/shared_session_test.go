package daemon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"jin/internal/session"
	"jin/internal/store"
	"jin/internal/testjin"
)

func TestTwoClientsShareOneAgentAndPausedQueue(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	testjin.OnPath(t)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"shared answer\"}}]}\n\ndata: [DONE]\n\n"))
	}))
	defer provider.Close()
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.AddProvider(store.ProviderEntry{ID: "p", Name: "p", Kind: "openai", BaseURL: provider.URL, APIKey: "k"}); err != nil {
		t.Fatal(err)
	}
	_ = db.SetActiveProvider("p")
	_ = db.SaveModel("m", "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager := session.NewManager(ctx, db, "v1", nil)
	manager.UseSharedQueue()
	defer manager.Shutdown()
	server := httptest.NewServer(routes("v1", manager, func() {}))
	defer server.Close()
	client := func() *Client {
		return &Client{Version: "v1", http: &http.Client{Transport: rewriteTransport{server.URL}}}
	}
	first, second := client(), client()
	snap, err := first.Create(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	id := snap.State.ID
	deadline := time.Now().Add(10 * time.Second)
	for !snap.State.Ready {
		if time.Now().After(deadline) {
			t.Fatal("session did not become ready")
		}
		time.Sleep(10 * time.Millisecond)
		snap, err = second.Snapshot(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := first.Command(ctx, Command{Action: "stop", Session: id}, nil); err != nil {
		t.Fatal(err)
	}
	send := Command{ID: "send-once", Action: "send", Session: id, Text: "from client two"}
	if err := second.Command(ctx, send, nil); err != nil {
		t.Fatal(err)
	}
	if err := first.Command(ctx, send, nil); err != nil {
		t.Fatal(err)
	}
	snap, _ = first.Snapshot(ctx, id)
	if snap.State.ReadOnly != 0 || snap.State.Queued != 1 || !snap.State.Paused {
		t.Fatalf("shared state: %+v", snap.State)
	}
	if err := first.Command(ctx, Command{Action: "resume", Session: id}, nil); err != nil {
		t.Fatal(err)
	}
	for {
		snap, err = second.Snapshot(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if !snap.State.Busy {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("session stayed busy")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(snap.Entries) != 2 || snap.Entries[1].Text != "shared answer" {
		t.Fatalf("entries: %+v", snap.Entries)
	}
	if len(manager.Live()) != 1 {
		t.Fatal("multiple agents were created")
	}
}
