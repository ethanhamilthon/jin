package ui

import (
	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
)

const (
	zoomIn  = "⤢"
	zoomOut = "⤡"
)

// zoomed is true while the focused pane fills the work view. The other panes
// stay open and keep running.
func (a *app) zoomed() bool {
	return a.zoom && len(paneLeaves(a.panes)) > 1 && a.focusedLeaf() != nil
}

// shownPanes is the part of the pane tree that is on the screen.
func (a *app) shownPanes() *paneNode {
	if a.zoomed() {
		return a.focusedLeaf()
	}
	return a.panes
}

func (a *app) toggleZoom() {
	switch {
	case a.zoomed():
		a.zoom = false
	case len(paneLeaves(a.panes)) < 2:
		a.active.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Zoom needs two or more panes"})
	default:
		a.zoom = true
	}
}

// zoomBlocked reports a layout action that zoom forbids and says why.
func (a *app) zoomBlocked() bool {
	if !a.zoomed() {
		return false
	}
	a.active.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Exit zoom first: /zoom"})
	return true
}

// zoomGlyph is the button on the top edge of the focused pane; panes show it
// only when there is more than one.
func (a *app) zoomGlyph(leaf *paneNode) string {
	switch {
	case leaf != a.focused:
		return ""
	case a.zoomed():
		return zoomOut
	case len(paneLeaves(a.panes)) > 1:
		return zoomIn
	}
	return ""
}

// zoomButton is the cell of the zoom glyph on the top edge of a pane.
func zoomButton(r paneRect) (x, y int) { return r.x + r.w - 4, r.y }

// zoomClick toggles zoom when a press lands on the zoom glyph of the
// focused pane.
func (a *app) zoomClick(ev *tcell.EventMouse, rects map[*paneNode]paneRect) bool {
	leaf := a.focusedLeaf()
	if leaf == nil || ev.Buttons()&tcell.ButtonPrimary == 0 || a.zoomGlyph(leaf) == "" {
		return false
	}
	x, y := ev.Position()
	bx, by := zoomButton(rects[leaf])
	if y != by || x < bx-1 || x > bx+1 {
		return false
	}
	a.toggleZoom()
	return true
}
