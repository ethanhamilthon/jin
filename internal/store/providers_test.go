package store

import (
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"jin/internal/paths"
	"jin/internal/provider"
)

func createOldV02DB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	file, err := paths.Global("jin.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		t.Fatal(err)
	}
	q := url.Values{}
	q.Add("_pragma", "journal_mode(WAL)")
	dsn := (&url.URL{Scheme: "file", Path: file, RawQuery: q.Encode()}).String()
	rawDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer rawDB.Close()

	oldSchema := `
	CREATE TABLE sessions (
		id TEXT PRIMARY KEY,
		path TEXT NOT NULL,
		model TEXT NOT NULL,
		effort TEXT NOT NULL,
		title TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		input_tokens INTEGER NOT NULL DEFAULT 0,
		output_tokens INTEGER NOT NULL DEFAULT 0,
		context_tokens INTEGER NOT NULL DEFAULT 0,
		cost REAL NOT NULL DEFAULT 0
	);
	CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT NOT NULL);
	`
	if _, err := rawDB.Exec(oldSchema); err != nil {
		t.Fatal(err)
	}
	inserts := `
	INSERT INTO settings (key, value) VALUES
		('provider.base_url', 'https://api.openai.com/v1'),
		('provider.api_key', 'sk-old-key'),
		('models.cache', '["gpt-4o","gpt-4o-mini"]'),
		('models.levels', '{"gpt-4o":["low","high"]}'),
		('models.scope', '["gpt-4o"]');
	INSERT INTO sessions (id, path, model, effort, title, created_at, updated_at)
		VALUES ('s1', '/tmp', 'gpt-4o', 'high', 'old session', 100, 200);
	`
	if _, err := rawDB.Exec(inserts); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestMigrationFromOldDB(t *testing.T) {
	home := createOldV02DB(t)
	t.Setenv("HOME", home)

	db, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	cfg, err := db.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ActiveProvider != "default" {
		t.Fatalf("active = %q, want 'default'", cfg.ActiveProvider)
	}
	if cfg.Provider.BaseURL != "https://api.openai.com/v1" || cfg.Provider.APIKey != "sk-old-key" || cfg.Provider.Kind != "openai" {
		t.Fatalf("cfg.Provider = %+v", cfg.Provider)
	}
	if len(cfg.Providers) != 1 || cfg.Providers[0].ID != "default" || cfg.Providers[0].Name != "default" {
		t.Fatalf("cfg.Providers = %+v", cfg.Providers)
	}

	// Idempotency: open again
	db2, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()
	cfg2, _ := db2.LoadConfig()
	if len(cfg2.Providers) != 1 || cfg2.ActiveProvider != "default" {
		t.Fatalf("after second open: providers=%+v active=%s", cfg2.Providers, cfg2.ActiveProvider)
	}

	// Old keys remain untouched in settings
	var oldURL, oldKey string
	_ = db.sql.QueryRow(`SELECT value FROM settings WHERE key='provider.base_url'`).Scan(&oldURL)
	_ = db.sql.QueryRow(`SELECT value FROM settings WHERE key='provider.api_key'`).Scan(&oldKey)
	if oldURL != "https://api.openai.com/v1" || oldKey != "sk-old-key" {
		t.Fatalf("old keys modified: url=%s key=%s", oldURL, oldKey)
	}

	// Fallback for models cache, levels, and scope
	cache, err := db.LoadModelsCache()
	if err != nil || !slices.Equal(cache, []string{"gpt-4o", "gpt-4o-mini"}) {
		t.Fatalf("LoadModelsCache: %v, %v", cache, err)
	}
	levels, err := db.LoadModelLevels()
	if err != nil || !slices.Equal(levels["gpt-4o"], []string{"low", "high"}) {
		t.Fatalf("LoadModelLevels: %v, %v", levels, err)
	}
	scope, err := db.LoadScope()
	if err != nil || !slices.Equal(scope, []string{"gpt-4o"}) {
		t.Fatalf("LoadScope: %v, %v", scope, err)
	}
	if !slices.Equal(cfg.Scope, []string{"gpt-4o"}) {
		t.Fatalf("cfg.Scope: %v", cfg.Scope)
	}

	// Old session column migration
	s1, ok, err := db.GetSession("s1")
	if err != nil || !ok || s1.Provider != "" || s1.Title != "old session" {
		t.Fatalf("old session: %+v, ok=%v, err=%v", s1, ok, err)
	}
}

func TestActiveProviderSwitchAndCRUD(t *testing.T) {
	db, _ := openTwo(t)

	p1 := ProviderEntry{ID: "p1", Name: "Provider One", Kind: "openai", BaseURL: "https://p1.example.com", APIKey: "key1"}
	p2 := ProviderEntry{ID: "p2", Name: "Provider Two", Kind: "anthropic", BaseURL: "https://p2.example.com", APIKey: "key2"}

	if err := db.AddProvider(p1); err != nil {
		t.Fatal(err)
	}
	if err := db.AddProvider(p2); err != nil {
		t.Fatal(err)
	}

	cfg, _ := db.LoadConfig()
	if cfg.ActiveProvider != "p1" {
		t.Fatalf("initial active = %q, want 'p1'", cfg.ActiveProvider)
	}
	if cfg.Provider.BaseURL != "https://p1.example.com" {
		t.Fatalf("provider baseURL = %q", cfg.Provider.BaseURL)
	}

	// Switch active
	if err := db.SetActiveProvider("p2"); err != nil {
		t.Fatal(err)
	}
	cfg, _ = db.LoadConfig()
	if cfg.ActiveProvider != "p2" || cfg.Provider.Kind != "anthropic" || cfg.Provider.BaseURL != "https://p2.example.com" {
		t.Fatalf("switched active config: %+v", cfg)
	}

	// Set invalid active
	if err := db.SetActiveProvider("unknown"); err == nil {
		t.Fatal("expected error for nonexistent provider")
	}

	// Delete active entry -> switches to remaining
	if err := db.DeleteProvider("p2"); err != nil {
		t.Fatal(err)
	}
	cfg, _ = db.LoadConfig()
	if cfg.ActiveProvider != "p1" || len(cfg.Providers) != 1 {
		t.Fatalf("after delete active: active=%s, providers=%+v", cfg.ActiveProvider, cfg.Providers)
	}

	// Delete last provider
	if err := db.DeleteProvider("p1"); err != nil {
		t.Fatal(err)
	}
	cfg, _ = db.LoadConfig()
	if cfg.ActiveProvider != "" || len(cfg.Providers) != 0 {
		t.Fatalf("after deleting all: active=%s, providers=%+v", cfg.ActiveProvider, cfg.Providers)
	}
}

func TestSaveProviderExistingCallers(t *testing.T) {
	db, _ := openTwo(t)

	// SaveProvider on fresh DB creates default entry
	err := db.SaveProvider(provider.Config{BaseURL: "https://fresh.example.com", APIKey: "sk-fresh"}, "m1", "low")
	if err != nil {
		t.Fatal(err)
	}
	cfg, _ := db.LoadConfig()
	if cfg.ActiveProvider != "default" || cfg.Model != "m1" || cfg.Effort != "low" {
		t.Fatalf("SaveProvider created: active=%s model=%s effort=%s", cfg.ActiveProvider, cfg.Model, cfg.Effort)
	}
	if cfg.Provider.BaseURL != "https://fresh.example.com" || cfg.Provider.APIKey != "sk-fresh" {
		t.Fatalf("cfg.Provider = %+v", cfg.Provider)
	}

	// Calling SaveProvider again updates active entry
	err = db.SaveProvider(provider.Config{BaseURL: "https://updated.example.com", APIKey: "sk-updated", Kind: "openai"}, "m2", "high")
	if err != nil {
		t.Fatal(err)
	}
	cfg, _ = db.LoadConfig()
	if len(cfg.Providers) != 1 || cfg.Provider.BaseURL != "https://updated.example.com" || cfg.Model != "m2" {
		t.Fatalf("SaveProvider update: %+v", cfg)
	}
}

func TestSessionProviderColumn(t *testing.T) {
	db, _ := openTwo(t)

	if err := db.Touch("s1", "/path", "m", "low", "title 1"); err != nil {
		t.Fatal(err)
	}
	s1, _, _ := db.GetSession("s1")
	if s1.Provider != "" {
		t.Fatalf("default session provider = %q, want empty", s1.Provider)
	}

	if err := db.TouchProvider("s2", "/path", "m", "low", "title 2", "custom-p"); err != nil {
		t.Fatal(err)
	}
	s2, _, _ := db.GetSession("s2")
	if s2.Provider != "custom-p" {
		t.Fatalf("s2 provider = %q, want 'custom-p'", s2.Provider)
	}

	// Touch without provider on existing session keeps existing provider
	if err := db.Touch("s2", "/path", "m-new", "high", "title 2"); err != nil {
		t.Fatal(err)
	}
	s2, _, _ = db.GetSession("s2")
	if s2.Provider != "custom-p" || s2.Model != "m-new" {
		t.Fatalf("s2 touched: %+v", s2)
	}

	if err := db.SetSessionProvider("s1", "new-p"); err != nil {
		t.Fatal(err)
	}
	s1, _, _ = db.GetSession("s1")
	if s1.Provider != "new-p" {
		t.Fatalf("s1 updated provider = %q", s1.Provider)
	}

	list, err := db.ListByPath("/path")
	if err != nil || len(list) != 2 {
		t.Fatalf("ListByPath: %v, %v", list, err)
	}
}

func TestPerProviderCacheAndScope(t *testing.T) {
	db, _ := openTwo(t)

	p1 := ProviderEntry{ID: "p1", Name: "P1", Kind: "openai", BaseURL: "http://p1", APIKey: "k1"}
	p2 := ProviderEntry{ID: "p2", Name: "P2", Kind: "openai", BaseURL: "http://p2", APIKey: "k2"}
	_ = db.AddProvider(p1)
	_ = db.AddProvider(p2)
	_ = db.SetActiveProvider("p1")

	_ = db.SaveModelsCacheFor("p1", []string{"p1-m1", "p1-m2"})
	_ = db.SaveModelsCacheFor("p2", []string{"p2-m1"})

	_ = db.SaveModelLevelsFor("p1", map[string][]string{"p1-m1": {"low"}})
	_ = db.SaveModelLevelsFor("p2", map[string][]string{"p2-m1": {"high"}})

	_ = db.SaveScopeFor("p1", []string{"p1-m1"})
	_ = db.SaveScopeFor("p2", []string{"p2-m1"})

	// Active is p1
	c1, _ := db.LoadModelsCache()
	if !slices.Equal(c1, []string{"p1-m1", "p1-m2"}) {
		t.Fatalf("p1 cache: %v", c1)
	}
	l1, _ := db.LoadModelLevels()
	if !slices.Equal(l1["p1-m1"], []string{"low"}) {
		t.Fatalf("p1 levels: %v", l1)
	}
	s1, _ := db.LoadScope()
	if !slices.Equal(s1, []string{"p1-m1"}) {
		t.Fatalf("p1 scope: %v", s1)
	}

	// Switch to p2
	_ = db.SetActiveProvider("p2")
	c2, _ := db.LoadModelsCache()
	if !slices.Equal(c2, []string{"p2-m1"}) {
		t.Fatalf("p2 cache: %v", c2)
	}
	l2, _ := db.LoadModelLevels()
	if !slices.Equal(l2["p2-m1"], []string{"high"}) {
		t.Fatalf("p2 levels: %v", l2)
	}
	s2, _ := db.LoadScope()
	if !slices.Equal(s2, []string{"p2-m1"}) {
		t.Fatalf("p2 scope: %v", s2)
	}

	// Direct LoadFor
	c1Direct, _ := db.LoadModelsCacheFor("p1")
	if !slices.Equal(c1Direct, []string{"p1-m1", "p1-m2"}) {
		t.Fatalf("p1 direct cache: %v", c1Direct)
	}
}
