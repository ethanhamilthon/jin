package ui

import (
	"strings"
	"testing"
)

func zoomTestApp(t *testing.T) *app {
	t.Helper()
	a := paneTestApp(t, 80, 30)
	if !a.splitPane(true) {
		t.Fatal("split was rejected")
	}
	stopPaneStartup(a)
	return a
}

func TestZoomNeedsTwoPanes(t *testing.T) {
	a := paneTestApp(t, 80, 30)
	a.toggleZoom()
	if a.zoomed() {
		t.Fatal("zoom turned on with one pane")
	}
}

func TestZoomShowsOnlyTheFocusedPane(t *testing.T) {
	a := zoomTestApp(t)
	focused := a.focusedLeaf()
	a.toggleZoom()
	if !a.zoomed() || a.shownPanes() != focused {
		t.Fatal("zoom does not show the focused pane")
	}
	w, h := a.screen.Size()
	rects := paneRects(a.shownPanes(), w, paneArea(a, w, h).h)
	if len(rects) != 1 || rects[focused].w != w {
		t.Fatalf("rects = %+v", rects)
	}
	a.toggleZoom()
	if a.zoomed() || len(paneLeaves(a.shownPanes())) != 2 {
		t.Fatal("zoom did not turn off")
	}
}

func TestZoomBlocksSplitCloseAndFocusChange(t *testing.T) {
	a := zoomTestApp(t)
	a.toggleZoom()
	focused := a.focusedLeaf()
	if a.splitPane(true) {
		t.Error("split worked while zoomed")
	}
	a.closePane()
	a.cyclePaneFocus()
	if len(paneLeaves(a.panes)) != 2 || a.focusedLeaf() != focused {
		t.Error("layout or focus changed while zoomed")
	}
	last := a.active.history[len(a.active.history)-1]
	if !strings.Contains(last.text, "Exit zoom first") {
		t.Errorf("last entry = %q", last.text)
	}
}

func TestZoomIsOffWhenOnePaneIsLeft(t *testing.T) {
	a := zoomTestApp(t)
	a.toggleZoom()
	a.zoom = true
	a.panes = a.focusedLeaf()
	if a.zoomed() {
		t.Fatal("zoom stays on with one pane")
	}
}

func TestZoomGlyphOnlyOnTheFocusedPane(t *testing.T) {
	a := zoomTestApp(t)
	for _, leaf := range paneLeaves(a.panes) {
		want := ""
		if leaf == a.focused {
			want = zoomIn
		}
		if got := a.zoomGlyph(leaf); got != want {
			t.Errorf("glyph = %q, want %q", got, want)
		}
	}
	a.toggleZoom()
	if a.zoomGlyph(a.focusedLeaf()) != zoomOut {
		t.Error("zoomed pane has no exit glyph")
	}
}
