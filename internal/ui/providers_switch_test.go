package ui

import (
	"strings"
	"testing"

	"jin/internal/store"
)

func TestProviderListKeepsRemovedSubscriptionUnavailable(t *testing.T) {
	a, _ := layoutApp(t)
	entry := store.ProviderEntry{ID: "legacy", Name: "Old subscription", Source: "cliproxy", Profile: "codex"}
	a.cfg.Providers = []store.ProviderEntry{entry}
	sel := a.showProviders("legacy")
	if len(sel.options) != 1 {
		t.Fatalf("options = %+v", sel.options)
	}
	row := sel.options[0]
	if !strings.Contains(row.detail, "Subscription support was removed. Connect an API provider.") {
		t.Fatalf("detail = %q", row.detail)
	}
	if len(row.choices) != 0 {
		t.Fatalf("removed record has enable choices: %v", row.choices)
	}
	if err := sel.submit("legacy"); err == nil {
		t.Fatal("removed record cannot become the default")
	}
	if err := sel.onChoice("legacy", 0); err != nil {
		t.Fatalf("disabled record choice returned error: %v", err)
	}
	if _, ok := sel.actions['d']; !ok {
		t.Fatal("removed record must remain deletable")
	}
}
