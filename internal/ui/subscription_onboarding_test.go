package ui

import (
	"jin/internal/store"
	"testing"
)

func TestSubscriptionSetupCanReachSettingsBeforeChoosingAModel(t *testing.T) {
	a, _ := layoutApp(t)
	entry := store.ProviderEntry{ID: "subscription-codex", Source: "cliproxy", Profile: "codex"}
	a.cfg = store.Config{Providers: []store.ProviderEntry{entry}, Provider: entry.Config(), ActiveProvider: entry.ID}
	if !a.onboarding() {
		t.Fatal("a subscription without a model must keep the first-run screen")
	}
	a.cfg.Model = "m"
	if a.onboarding() {
		t.Fatal("a chosen model opens the chat")
	}
}
