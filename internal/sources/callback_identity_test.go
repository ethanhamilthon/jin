package sources

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"jin/internal/provider"
	"jin/internal/store"
)

func okServer(calls *atomic.Int32) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n"))
	}))
}

func stream(c *provider.Client, t *testing.T) error {
	_, err := c.Stream(t.Context(), "model", "", nil, json.RawMessage("[]"), func(provider.StreamEvent) {})
	return err
}

func TestClientRejectsChangedURLWithoutRequest(t *testing.T) {
	db := sourceDB(t)
	var oldCalls, newCalls atomic.Int32
	oldServer, newServer := okServer(&oldCalls), okServer(&newCalls)
	defer oldServer.Close()
	defer newServer.Close()
	entry := store.ProviderEntry{ID: "p", Name: "P", BaseURL: oldServer.URL, APIKey: "fixture"}
	if err := db.AddProvider(entry); err != nil {
		t.Fatal(err)
	}
	client := Client(db, entry)
	changed := entry
	changed.BaseURL = newServer.URL
	if err := db.SaveProviders([]store.ProviderEntry{changed}, "p"); err != nil {
		t.Fatal(err)
	}
	err := stream(client, t)
	if err == nil || !strings.Contains(err.Error(), "changed") || oldCalls.Load() != 0 || newCalls.Load() != 0 {
		t.Fatalf("err=%v old=%d new=%d", err, oldCalls.Load(), newCalls.Load())
	}
}

func TestClientRejectsChangedKindWithoutRequest(t *testing.T) {
	db := sourceDB(t)
	var calls atomic.Int32
	server := okServer(&calls)
	defer server.Close()
	entry := store.ProviderEntry{ID: "p", Name: "P", Kind: provider.KindOpenAI, BaseURL: server.URL, APIKey: "fixture"}
	if err := db.AddProvider(entry); err != nil {
		t.Fatal(err)
	}
	client := Client(db, entry)
	changed := entry
	changed.Kind = provider.KindResponses
	if err := db.SaveProviders([]store.ProviderEntry{changed}, "p"); err != nil {
		t.Fatal(err)
	}
	if err := stream(client, t); err == nil || calls.Load() != 0 {
		t.Fatalf("err=%v calls=%d", err, calls.Load())
	}
}

func TestClientEnableToggleAllowsRequestAgain(t *testing.T) {
	db := sourceDB(t)
	var calls atomic.Int32
	server := okServer(&calls)
	defer server.Close()
	entry := store.ProviderEntry{ID: "p", Name: "P", BaseURL: server.URL, APIKey: "fixture"}
	if err := db.AddProvider(entry); err != nil {
		t.Fatal(err)
	}
	client := Client(db, entry)
	_ = db.SetProviderEnabled("p", false)
	if err := stream(client, t); err == nil || calls.Load() != 0 {
		t.Fatalf("disabled err=%v calls=%d", err, calls.Load())
	}
	_ = db.SetProviderEnabled("p", true)
	if err := stream(client, t); err != nil || calls.Load() != 1 {
		t.Fatalf("enabled err=%v calls=%d", err, calls.Load())
	}
}
