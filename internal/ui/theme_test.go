package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestThemesAreCompleteAndDistinct(t *testing.T) {
	if len(themes) != 10 || themes[0].name != "Jin Original" {
		t.Fatalf("themes = %d, first %q", len(themes), themes[0].name)
	}
	seen := map[string]bool{}
	for _, th := range themes {
		if seen[th.name] || th.fg == th.bg || th.status == 0 {
			t.Errorf("theme %q is a duplicate or incomplete", th.name)
		}
		seen[th.name] = true
	}
}

func TestThemePreviewAndCancel(t *testing.T) {
	db, _ := openFoldDB(t)
	a, screen := layoutApp(t)
	a.store = db
	t.Cleanup(func() { applyTheme(themes[0]) })
	a.openThemeFlow()
	a.selectorKey(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone))
	if colorBG != rgb(themes[1].bg) {
		t.Fatalf("preview did not apply %s", themes[1].name)
	}
	a.selectorKey(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone))
	if colorBG != rgb(themes[0].bg) {
		t.Fatal("Esc must restore the theme")
	}
	a.openThemeFlow()
	a.selectorKey(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone))
	a.selectorKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	if cfg, _ := db.LoadConfig(); cfg.Theme != themes[1].name {
		t.Fatalf("saved theme = %q", cfg.Theme)
	}
	a.draw()
	if _, style, _ := screen.Get(0, 23); style.GetBackground() != rgb(themes[1].status) {
		t.Fatal("status bar did not take the theme color")
	}
}
