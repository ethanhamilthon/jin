package headless

import (
	"testing"

	"jin/internal/provider"
	"jin/internal/store"
)

func TestUseSavedPinsProviderIdentityWhenConfigsMatch(t *testing.T) {
	cfg := store.Config{
		ActiveProvider: "codex",
		Provider:       provider.Config{Kind: provider.KindOpenAI, BaseURL: "http://proxy.test", APIKey: "fixture"},
		Providers: []store.ProviderEntry{
			{ID: "codex", Name: "Codex", Kind: provider.KindOpenAI, BaseURL: "http://proxy.test", APIKey: "fixture"},
			{ID: "claude", Name: "Claude", Kind: provider.KindOpenAI, BaseURL: "http://proxy.test", APIKey: "fixture"},
		},
	}
	id, err := useSaved(&cfg, "claude")
	if err != nil || id != "claude" || cfg.ActiveProvider != "claude" {
		t.Fatalf("id=%q active=%q err=%v", id, cfg.ActiveProvider, err)
	}
}
