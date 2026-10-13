package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"

	"jin/internal/provider"
	"jin/internal/store"
)

func TestOnboardingGatesTheChat(t *testing.T) {
	a, _ := layoutApp(t)
	a.cfg = store.Config{}
	if !a.onboarding() {
		t.Fatal("no provider must show the first-run screen")
	}
	press := func(key tcell.Key, text string) { a.handleEvent(tcell.NewEventKey(key, text, tcell.ModNone)) }
	if len(a.onboardingChoices()) != 2 {
		t.Fatalf("setup must expose two API choices, got %d", len(a.onboardingChoices()))
	}
	press(tcell.KeyEnter, "")
	if a.sel == nil || a.sel.title != "Name" || a.sel.query[0] != "o" {
		t.Fatalf("Enter must start OpenAI-compatible setup: %+v", a.sel)
	}
	a.sel = nil
	press(tcell.KeyRune, "2")
	if a.sel == nil || a.sel.title != "Name" || a.sel.query[0] != "a" {
		t.Fatalf("2 must start Anthropic setup: %+v", a.sel)
	}
	a.handleEvent(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone))
	if a.sel != nil || !a.onboarding() {
		t.Fatal("Esc must return to the first-run screen")
	}
	a.cfg.Provider = provider.Config{BaseURL: "http://x", APIKey: "k"}
	if !a.onboarding() {
		t.Fatal("a provider without a model must still show the first-run screen")
	}
	a.cfg.Model = "m"
	if a.onboarding() {
		t.Fatal("provider and model open the chat")
	}
	entry := store.ProviderEntry{ID: "old", Source: "cliproxy"}
	a.cfg = store.Config{Providers: []store.ProviderEntry{entry}, ActiveProvider: entry.ID, Provider: entry.Config(), Model: "old-model"}
	if !a.onboarding() {
		t.Fatal("removed subscription default must require explicit API recovery")
	}
	if err := a.activateProvider(entry.ID); err == nil {
		t.Fatal("removed subscription cannot become the selected provider")
	}
	a.cfg = store.Config{}
	a.handleEvent(tcell.NewEventKey(tcell.KeyCtrlC, "", tcell.ModCtrl))
	if !a.quit {
		t.Fatal("Ctrl+C quits the first-run screen")
	}
}
