package sources

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"jin/internal/provider"
	"jin/internal/store"
)

func TestRemovedProviderRequestIsBlockedWithoutChangingRecord(t *testing.T) {
	db := sourceDB(t)
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()
	if err := db.SaveProviders([]store.ProviderEntry{{ID: "subscription-codex", Name: "Codex", Kind: "openai", BaseURL: server.URL, APIKey: "fixture", Source: "cliproxy", Profile: "codex"}}, "subscription-codex"); err != nil {
		t.Fatal(err)
	}
	before, err := db.Setting("providers")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := db.Provider("subscription-codex")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Client(db, entry).Stream(t.Context(), "codex/model", "", nil, json.RawMessage("[]"), func(provider.StreamEvent) {})
	if err == nil || !strings.Contains(err.Error(), "support was removed") || calls != 0 {
		t.Fatalf("request error=%v calls=%d", err, calls)
	}
	got, err := db.Provider(entry.ID)
	after, readErr := db.Setting("providers")
	if err != nil || readErr != nil || got.Source != "cliproxy" || got.Profile != "codex" || got.APIKey != "fixture" || !got.Config().Ready() || before != after {
		t.Fatalf("legacy metadata changed: %+v bytesEqual=%t err=%v readErr=%v", got, before == after, err, readErr)
	}
}

func TestCatalogExcludesEnabledRemovedProviderWithoutRequest(t *testing.T) {
	db := sourceDB(t)
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"data":[{"id":"legacy-model"}]}`))
	}))
	defer server.Close()
	if err := db.AddProvider(store.ProviderEntry{ID: "legacy", Name: "Legacy", Kind: "openai", BaseURL: server.URL, APIKey: "fixture", Source: "cliproxy", Profile: "codex"}); err != nil {
		t.Fatal(err)
	}
	catalog, err := List(t.Context(), db, false)
	if err != nil || len(catalog.Models) != 0 || calls != 0 {
		t.Fatalf("catalog=%+v calls=%d err=%v", catalog, calls, err)
	}
}
