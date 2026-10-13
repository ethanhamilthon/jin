package store

import (
	"strings"
	"testing"
)

func TestProviderMigrationAndRemovedGuards(t *testing.T) {
	db := openTest(t)
	_, err := db.sql.Exec(`INSERT INTO settings(key,value) VALUES(?,?)`, keyProviders,
		`[{"id":"chat","kind":"openai","api_key":"keep"},{"id":"resp","kind":"responses"},{"id":"empty"},{"id":"anth","kind":"anthropic"},{"id":"old","kind":"openai","source":"cliproxy","profile":"codex"}]`)
	if err != nil {
		t.Fatal(err)
	}
	list, _, err := db.LoadProviders()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"chat": "openai", "resp": "responses", "empty": "openai", "anth": "anthropic", "old": "openai"}
	for _, p := range list {
		if p.ID == "old" {
			if !p.Disabled || !p.Removed() || p.Profile != "codex" {
				t.Fatalf("removed entry: %+v", p)
			}
		} else if p.ID == "anth" {
			if p.Kind != "anthropic" || p.Protocol != "" {
				t.Fatalf("anthropic changed: %+v", p)
			}
		} else if p.Kind != "auto" || p.Protocol != want[p.ID] {
			t.Fatalf("migration: %+v", p)
		}
	}
	if err := db.SetProviderEnabled("old", true); err == nil {
		t.Fatal("enabled removed provider")
	}
	if err := db.SetActiveProvider("old"); err == nil {
		t.Fatal("selected removed default")
	}
	if err := db.SelectModel("old", "codex/model", "", true); err == nil {
		t.Fatal("selected removed model")
	}
	if err := db.SetProviderEnabled("chat", false); err != nil {
		t.Fatal(err)
	}
	if err := db.SetProviderEnabled("chat", true); err != nil {
		t.Fatal(err)
	}
	again, _, _ := db.LoadProviders()
	if again[0].Protocol != "openai" {
		t.Fatalf("migration not idempotent: %+v", again[0])
	}
}

func TestProtocolPersistenceScopeStaleAndInvalidation(t *testing.T) {
	db := openTest(t)
	if err := db.AddProvider(ProviderEntry{ID: "p", Kind: "auto", BaseURL: "HTTPS://Example.test/v1/", APIKey: "secret"}); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveProtocol("p", 0, "https://example.test/v1", "m", "openai"); err != nil {
		t.Fatal(err)
	}
	if got, err := db.LoadProtocol("p", 0, "https://EXAMPLE.test/v1/", "m"); err != nil || got != "openai" {
		t.Fatalf("load=%q err=%v", got, err)
	}
	if got, err := db.LoadProtocol("other", 0, "https://example.test/v1", "m"); err != nil || got != "" {
		t.Fatalf("scope=%q err=%v", got, err)
	}
	if err := db.AddProvider(ProviderEntry{ID: "p", Kind: "auto", BaseURL: "https://changed.test", APIKey: "secret"}); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveProtocol("p", 0, "https://example.test/v1", "m", "responses"); err == nil {
		t.Fatal("accepted stale save")
	}
	if got, _ := db.LoadProtocol("p", 1, "https://changed.test", "m"); got != "" {
		t.Fatalf("old choice survived URL change: %q", got)
	}
	if err := db.AddProvider(ProviderEntry{ID: "q", Kind: "auto", BaseURL: "https://changed.test"}); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveProtocol("q", 0, "https://changed.test", "m", "invalid"); err == nil || !strings.Contains(err.Error(), "protocol") {
		t.Fatalf("invalid protocol error=%v", err)
	}
	if err := db.DeleteProvider("p"); err != nil {
		t.Fatal(err)
	}
}

func TestProtocolRevertedIdentityStaysInvalid(t *testing.T) {
	db := openTest(t)
	a := ProviderEntry{ID: "p", Kind: "auto", BaseURL: "https://a.test"}
	b := ProviderEntry{ID: "p", Kind: "auto", BaseURL: "https://b.test"}
	for _, e := range []ProviderEntry{a} {
		if err := db.AddProvider(e); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.SaveProtocol("p", 0, a.BaseURL, "m", "openai"); err != nil {
		t.Fatal(err)
	}
	for _, e := range []ProviderEntry{b, a} {
		if err := db.AddProvider(e); err != nil {
			t.Fatal(err)
		}
	}
	if got, _ := db.LoadProtocol("p", 2, a.BaseURL, "m"); got != "" {
		t.Fatalf("resurrected %q", got)
	}
	if err := db.SaveProtocol("p", 0, a.BaseURL, "m", "responses"); err == nil {
		t.Fatal("stale epoch save accepted")
	}
}
