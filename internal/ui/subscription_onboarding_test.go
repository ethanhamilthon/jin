package ui

import (
	"testing"

	"jin/internal/store"
)

func TestRemovedSubscriptionDefaultRequiresAPIRecovery(t *testing.T) {
	a, _ := layoutApp(t)
	entry := store.ProviderEntry{ID: "subscription-codex", Source: "cliproxy", Profile: "codex"}
	a.cfg = store.Config{Providers: []store.ProviderEntry{entry}, Provider: entry.Config(), ActiveProvider: entry.ID, Model: "legacy-model"}
	if !a.onboarding() {
		t.Fatal("a removed subscription default requires explicit API recovery")
	}
	if err := a.activateProvider(entry.ID); err == nil {
		t.Fatal("a removed subscription cannot be selected")
	}
}
