package ui

import (
	"jin/internal/paths"
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v3"

	"jin/internal/provider"
	"jin/internal/store"
)

func TestOnboardingGatesTheChat(t *testing.T) {
	a, _ := layoutApp(t)
	installFakeProxy(t)
	a.cfg = store.Config{}
	if !a.onboarding() {
		t.Fatal("no provider must show the first-run screen")
	}
	press := func(key tcell.Key, text string) { a.handleEvent(tcell.NewEventKey(key, text, tcell.ModNone)) }
	press(tcell.KeyEnter, "")
	if a.onboardMode != modeAPI || a.sel != nil {
		t.Fatalf("the first choice must be API: %q %+v", a.onboardMode, a.sel)
	}
	press(tcell.KeyEscape, "")
	if a.onboardMode != "" {
		t.Fatal("Esc must return to the API or subscription choice")
	}
	press(tcell.KeyRune, "2")
	if a.onboardMode != modeSubscription || len(a.onboardingChoices()) != 3 {
		t.Fatalf("2 must open the three subscriptions: %q", a.onboardMode)
	}
	press(tcell.KeyEscape, "")
	press(tcell.KeyEnter, "")
	press(tcell.KeyDown, "")
	press(tcell.KeyEnter, "")
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

// installFakeProxy gives the test its own home with a CLIProxyAPI that looks
// installed, so no test reads the real data folder or downloads anything.
func installFakeProxy(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, paths.DirName(), "cliproxyapi")
	binary := filepath.Join(root, "versions", "v8.0.23", "cli-proxy-api")
	if err := os.MkdirAll(filepath.Dir(binary), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binary, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "installed.json"), []byte(`{"version":"v8.0.23"}`), 0o644); err != nil {
		t.Fatal(err)
	}
}
