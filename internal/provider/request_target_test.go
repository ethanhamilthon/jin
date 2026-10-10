package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestRequestTargetReleasesLeaseOnEveryAttempt(t *testing.T) {
	var calls, released atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(503)
			return
		}
		w.Write([]byte(okStream))
	}))
	defer server.Close()
	cfg := Config{BaseURL: server.URL, APIKey: "fixture"}
	client := NewClient(cfg)
	client.SetTarget(func(context.Context, Config, string, []byte) (Config, func(), error) {
		return cfg, func() { released.Add(1) }, nil
	}, "")
	if _, _, err := streamOnce(t, client); err != nil {
		t.Fatal(err)
	}
	if released.Load() != 2 {
		t.Fatalf("released=%d", released.Load())
	}
}

func TestModelPrefixFiltersTheCatalog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{"id": "claude/shared"}, {"id": "codex/shared"}}})
	}))
	defer server.Close()
	client := NewClient(Config{BaseURL: server.URL, APIKey: "fixture"})
	client.SetTarget(nil, "codex/")
	list, err := client.Models(t.Context())
	if err != nil || len(list) != 1 || list[0] != "codex/shared" {
		t.Fatalf("list=%v err=%v", list, err)
	}
}
