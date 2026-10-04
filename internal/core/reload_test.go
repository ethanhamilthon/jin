package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"jin/internal/provider"
	"jin/internal/tools"
)

func startReloadRun(t *testing.T, agent *Agent) (context.CancelFunc, chan Request, chan Update, <-chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	prompts, updates, done := make(chan Request, 8), make(chan Update, 64), make(chan struct{})
	go func() { agent.Run(ctx, nil, prompts, updates); close(done) }()
	t.Cleanup(func() { cancel(); <-done })
	return cancel, prompts, updates, done
}

func controlledReloadAgent(t *testing.T) (*Agent, <-chan string, func()) {
	t.Helper()
	systems, release := make(chan string, 8), make(chan struct{})
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []provider.Message `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		systems <- body.Messages[0].Content
		if calls.Add(1) == 1 {
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n")
	}))
	t.Cleanup(func() { server.Close() })
	agent := NewAgent(provider.NewClient(provider.Config{BaseURL: server.URL, APIKey: "k"}), "old", tools.NewRegistry())
	var once atomic.Bool
	return agent, systems, func() {
		if once.CompareAndSwap(false, true) {
			close(release)
		}
	}
}

func waitReloadQueued(t *testing.T, agent *Agent) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for len(agent.reloads) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if len(agent.reloads) == 0 {
		t.Fatal("prompt reload was not queued")
	}
}
