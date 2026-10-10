package ui

import (
	"jin/internal/store"
	"testing"
)

func TestSubscriptionSetupCanReachSettingsBeforeChoosingAModel(t *testing.T) {
	a, _ := layoutApp(t)
	entry := store.ProviderEntry{ID: "subscription-codex", Source: "cliproxy", Profile: "codex"}
	a.cfg = store.Config{Providers: []store.ProviderEntry{entry}, Provider: entry.Config(), ActiveProvider: entry.ID}
	if a.onboarding() {
		t.Fatal("subscription setup cannot reach install/sign-in settings")
	}
	a.startOnboarding()
	if a.sel != nil {
		t.Fatal("signed-out subscription opened model discovery automatically")
	}
}
