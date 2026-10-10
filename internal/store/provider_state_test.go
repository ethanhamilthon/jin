package store

import (
	"testing"
)

func TestProviderSwitchKeepsDefaultAndCredentials(t *testing.T) {
	db := openTest(t)
	p := ProviderEntry{ID: "p", Name: "Provider", BaseURL: "http://example.test", APIKey: "fixture"}
	if err := db.AddProvider(p); err != nil {
		t.Fatal(err)
	}
	if err := db.SetProviderEnabled("p", false); err != nil {
		t.Fatal(err)
	}
	cfg, err := db.LoadConfig()
	if err != nil || cfg.ActiveProvider != "p" || cfg.Provider.Ready() {
		t.Fatalf("default/ready: %+v %v", cfg, err)
	}
	saved, err := db.Provider("p")
	if err != nil || !saved.Disabled || saved.APIKey != "fixture" {
		t.Fatalf("entry: %+v %v", saved, err)
	}
	if err := db.SelectModel("p", "model", "", true); err == nil {
		t.Fatal("disabled model selected")
	}
	if err := db.SetProviderEnabled("p", true); err != nil {
		t.Fatal(err)
	}
	cfg, _ = db.LoadConfig()
	if !cfg.Provider.Ready() {
		t.Fatal("provider not restored")
	}
}

func TestProviderLegacyAndManagedMetadata(t *testing.T) {
	list := parseProviders(`[{"id":"old","base_url":"http://example.test","api_key":"fixture"}]`)
	if len(list) != 1 || list[0].Disabled || !list[0].Config().Ready() {
		t.Fatal("legacy provider changed")
	}
	p := ProviderEntry{ID: "managed", Source: "cliproxy", Profile: "codex"}
	if !p.Config().Managed || !p.Config().Ready() {
		t.Fatal("managed provider metadata missing")
	}
}
