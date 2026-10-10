package ui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v3"

	"jin/internal/sources"
)

func settingsApp(t *testing.T) *app {
	t.Helper()
	a := startingApp(t)
	cfg, err := a.store.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	a.cfg = cfg
	a.loads = make(chan loadResult, 4)
	return a
}

func loadStep(t *testing.T, a *app) {
	t.Helper()
	select {
	case result := <-a.loads:
		a.receiveLoad(result)
	case <-time.After(5 * time.Second):
		t.Fatal("list did not load")
	}
}

func TestSessionTitlesRowOpensFromSettings(t *testing.T) {
	a := settingsApp(t)
	a.openSettingsFlow()
	a.sel.selectValue("Session titles")
	a.submitSelector()
	if a.sel == nil || a.sel.title != "Session titles" {
		t.Fatalf("the Session titles list did not open: %+v", a.sel)
	}
}

func TestSessionTitlesSaveTheSwitchAndTheAfterCount(t *testing.T) {
	a := settingsApp(t)
	a.openTitleFlow()
	a.sel.selectValue("refresh")
	press(a, tcell.KeyLeft)
	if cfg, _ := a.store.LoadConfig(); !cfg.Title.Refresh {
		t.Fatal("Left should switch Rename at message 4 on")
	}
	a.sel.selectValue("after")
	a.submitSelector()
	a.sel.query = clusters("3")
	a.submitSelector()
	if cfg, _ := a.store.LoadConfig(); cfg.Title.After != 3 {
		t.Fatalf("after = %d, want 3", cfg.Title.After)
	}
	if a.sel == nil || a.sel.title != "Session titles" {
		t.Fatal("saving the count should show the list again")
	}
}

func TestSessionTitlesRejectACountOutOfRange(t *testing.T) {
	a := settingsApp(t)
	a.openTitleFlow()
	a.sel.selectValue("after")
	a.submitSelector()
	a.sel.query = clusters("51")
	a.submitSelector()
	if a.sel == nil || a.sel.err == "" {
		t.Fatal("51 messages must be refused with a message")
	}
	if cfg, _ := a.store.LoadConfig(); cfg.Title.After != 1 {
		t.Fatalf("after changed to %d", cfg.Title.After)
	}
}

func TestSessionTitleModelPickSavesProviderAndModelThenResets(t *testing.T) {
	var hitsA, hitsB int
	a := twoProviderApp(t, modelServer(t, &hitsA, "a1"), modelServer(t, &hitsB, "b1", "b2"))
	a.loads = make(chan loadResult, 4)
	a.openTitleFlow()
	a.sel.selectValue("model")
	a.submitSelector()
	loadStep(t, a)
	a.sel.selectValue(sources.Key("b", "b2"))
	press(a, tcell.KeyEnter)
	loadStep(t, a)
	press(a, tcell.KeyEnter)
	if cfg, _ := a.store.LoadConfig(); cfg.Title.Provider != "b" || cfg.Title.Model != "b2" {
		t.Fatalf("title model = %s/%s", cfg.Title.Provider, cfg.Title.Model)
	}
	a.sel.selectValue("model")
	a.handleEvent(tcell.NewEventKey(tcell.KeyRune, "r", tcell.ModNone))
	if cfg, _ := a.store.LoadConfig(); cfg.Title.Model != "" || cfg.Title.Provider != "" {
		t.Fatalf("r should give back the session model, got %s/%s", cfg.Title.Provider, cfg.Title.Model)
	}
}

func TestSessionTitlePromptAsksForAnEditorFirst(t *testing.T) {
	a := settingsApp(t)
	a.cfg.Editor = ""
	a.openTitleFlow()
	a.sel.selectValue("prompt")
	a.submitSelector()
	if a.sel == nil || a.sel.title != "Editor" {
		t.Fatalf("without an editor the prompt must ask for one, list %+v", a.sel)
	}
}
