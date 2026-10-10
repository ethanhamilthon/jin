package sources

import (
	"context"
	"fmt"
	"jin/internal/store"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func sourceDB(t *testing.T) *store.DB {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCatalogPreservesDuplicateNamesAndScopes(t *testing.T) {
	db := sourceDB(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		fmt.Fprint(w, `{"data":[{"id":"shared"},{"id":"other"}]}`)
	}))
	defer server.Close()
	for _, id := range []string{"a", "b"} {
		if err := db.AddProvider(store.ProviderEntry{ID: id, Name: id, BaseURL: server.URL, APIKey: "fixture"}); err != nil {
			t.Fatal(err)
		}
	}
	_ = db.SaveScopeFor("a", []string{"shared"})
	catalog, err := List(t.Context(), db, false)
	if err != nil || len(catalog.Models) != 3 {
		t.Fatalf("catalog=%+v err=%v", catalog, err)
	}
	if Key("a", "shared") == Key("b", "shared") {
		t.Fatal("model keys collided")
	}
	_ = db.SetProviderEnabled("b", false)
	calls.Store(0)
	catalog, err = List(context.Background(), db, false)
	if err != nil || len(catalog.Models) != 1 || calls.Load() != 1 {
		t.Fatalf("disabled discovery=%+v calls=%d err=%v", catalog, calls.Load(), err)
	}
}

func TestCatalogKeepsHealthyProviderOnFailure(t *testing.T) {
	db := sourceDB(t)
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"data":[{"id":"ok"}]}`) }))
	defer good.Close()
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "unavailable", 503) }))
	defer bad.Close()
	for id, address := range map[string]string{"good": good.URL, "bad": bad.URL} {
		_ = db.AddProvider(store.ProviderEntry{ID: id, Name: id, BaseURL: address, APIKey: "fixture"})
	}
	result, err := List(t.Context(), db, false)
	if err != nil || len(result.Models) != 1 || len(result.Errors) != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
