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
	a.handleEvent(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone))
	a.handleEvent(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	if a.sel == nil || a.sel.title != "Name" || a.sel.query[0] != "o" {
		t.Fatalf("Enter must start the Chat Completions setup: %+v", a.sel)
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
	a.cfg = store.Config{}
	a.handleEvent(tcell.NewEventKey(tcell.KeyCtrlC, "", tcell.ModCtrl))
	if !a.quit {
		t.Fatal("Ctrl+C quits the first-run screen")
	}
}
