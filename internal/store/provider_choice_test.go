package store

import (
	"sync"
	"testing"
)

func TestProviderSwitchesDoNotOverwriteEachOther(t *testing.T) {
	one, two := openTwo(t)
	for _, id := range []string{"a", "b"} {
		_ = one.AddProvider(ProviderEntry{ID: id, Name: id, BaseURL: "http://example.test", APIKey: "fixture"})
	}
	var group sync.WaitGroup
	for i, db := range []*DB{one, two} {
		group.Add(1)
		go func(index int, db *DB) {
			defer group.Done()
			id := []string{"a", "b"}[index]
			if err := db.SetProviderEnabled(id, false); err != nil {
				t.Error(err)
			}
		}(i, db)
	}
	group.Wait()
	list, _, err := one.LoadProviders()
	if err != nil || len(list) != 2 || !list[0].Disabled || !list[1].Disabled {
		t.Fatalf("list=%+v err=%v", list, err)
	}
}

func TestChoicePersistsProviderModelAndScopedEffort(t *testing.T) {
	db := openTest(t)
	for _, id := range []string{"a", "b"} {
		_ = db.AddProvider(ProviderEntry{ID: id, Name: id, BaseURL: "http://example.test", APIKey: "fixture"})
	}
	_ = db.TouchProvider("s", "", "shared", "", "title", "a")
	if err := db.ChooseModel("s", "b", "shared", "high", true); err != nil {
		t.Fatal(err)
	}
	cfg, _ := db.LoadConfig()
	record, _, _ := db.GetSession("s")
	if cfg.ActiveProvider != "b" || cfg.Model != "shared" || record.Provider != "b" || record.Effort != "high" {
		t.Fatalf("choice not persisted: active=%s model=%s session=%+v", cfg.ActiveProvider, cfg.Model, record)
	}
	if err := db.RememberEffort("a", "shared", "low"); err != nil {
		t.Fatal(err)
	}
	cfg, _ = db.LoadConfig()
	if cfg.ModelEfforts[EffortKey("b", "shared")] != "high" || cfg.ModelEfforts[EffortKey("a", "shared")] != "low" {
		t.Fatal("model efforts collided")
	}
}
