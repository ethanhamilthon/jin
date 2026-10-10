package sources

import (
	"encoding/json"
	"fmt"
	"jin/internal/provider"
	"jin/internal/store"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDisablingDoesNotCutTheCurrentStream(t *testing.T) {
	db := sourceDB(t)
	started, finish := make(chan struct{}), make(chan struct{})
	var count atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"ok\"}}]}\n\n")
		w.(http.Flusher).Flush()
		close(started)
		<-finish
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()
	entry := store.ProviderEntry{ID: "p", Name: "P", BaseURL: server.URL, APIKey: "fixture"}
	_ = db.AddProvider(entry)
	client := Client(db, entry)
	done := make(chan error, 1)
	go func() {
		_, err := client.Stream(t.Context(), "model", "", nil, json.RawMessage("[]"), func(provider.StreamEvent) {})
		done <- err
	}()
	<-started
	if err := db.SetProviderEnabled("p", false); err != nil {
		t.Fatal(err)
	}
	close(finish)
	if err := <-done; err != nil {
		t.Fatalf("current stream interrupted: %v", err)
	}
	_, err := client.Stream(t.Context(), "model", "", nil, json.RawMessage("[]"), func(provider.StreamEvent) {})
	if err == nil || !strings.Contains(err.Error(), "disabled") || count.Load() != 1 {
		t.Fatalf("next stream=%v count=%d", err, count.Load())
	}
}

func TestDisableAlsoBlocksRetry(t *testing.T) {
	db := sourceDB(t)
	var count atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		_ = db.SetProviderEnabled("p", false)
		http.Error(w, "retry", 503)
	}))
	defer server.Close()
	entry := store.ProviderEntry{ID: "p", Name: "P", BaseURL: server.URL, APIKey: "fixture"}
	_ = db.AddProvider(entry)
	_, err := Client(db, entry).Stream(t.Context(), "model", "", nil, json.RawMessage("[]"), func(provider.StreamEvent) {})
	if err == nil || !strings.Contains(err.Error(), "disabled") || count.Load() != 1 {
		t.Fatalf("retry=%v count=%d", err, count.Load())
	}
}
