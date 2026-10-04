package ui

import (
	"path/filepath"

	"github.com/gdamore/tcell/v3"
)

func (a *app) drawPanes(height, width int) {
	if height <= 0 || a.panes == nil {
		return
	}
	rects := paneRects(a.panes, width, height)
	for _, leaf := range paneLeaves(a.panes) {
		r := rects[leaf]
		a.drawPaneFrame(leaf, r)
		a.drawPaneTimeline(leaf, r)
	}
}

func (a *app) drawPaneFrame(leaf *paneNode, r paneRect) {
	drawFrame(a.screen, r, paneTitle(leaf.session), leaf == a.focused)
}

// drawFrame draws a bordered rectangle with its label on the top border. The
// focused frame uses the theme accent; the others stay muted. The label of
// the focused frame sits on the primary color.
func drawFrame(screen tcell.Screen, r paneRect, label string, focused bool) {
	if r.w < 2 || r.h < 2 {
		return
	}
	style, labelStyle := muted, muted
	if focused {
		style, labelStyle = accent.Bold(true), paneLabel
	}
	put(screen, r.x, r.y, "┌", style)
	put(screen, r.x+r.w-1, r.y, "┐", style)
	put(screen, r.x, r.y+r.h-1, "└", style)
	put(screen, r.x+r.w-1, r.y+r.h-1, "┘", style)
	for x := r.x + 1; x < r.x+r.w-1; x++ {
		put(screen, x, r.y, "─", style)
		put(screen, x, r.y+r.h-1, "─", style)
	}
	for y := r.y + 1; y < r.y+r.h-1; y++ {
		put(screen, r.x, y, "│", style)
		put(screen, r.x+r.w-1, y, "│", style)
	}
	if r.w > 6 && label != "" {
		put(screen, r.x+2, r.y, " "+truncate(label, r.w-5)+" ", labelStyle)
	}
}

func paneTitle(s *chatSession) string {
	if s == nil {
		return ""
	}
	project := filepath.Base(filepath.Clean(s.path))
	if project == "." || project == string(filepath.Separator) {
		project = "project"
	}
	title := s.title
	if title == "" {
		title = "new session"
	}
	return project + " · " + title
}

func (a *app) drawPaneTimeline(leaf *paneNode, r paneRect) {
	if leaf.session == nil || r.w < 3 || r.h < 3 {
		return
	}
	s := leaf.session
	inner := paneRect{x: r.x + 1, y: r.y + 1, w: r.w - 2, h: r.h - 2}
	if s.width != inner.w {
		s.selection = textSelection{}
	}
	s.resize(max(1, inner.w))
	rows := append(s.rows[:len(s.rows):len(s.rows)], a.tailRows(s)...)
	s.scroll = min(max(0, s.scroll), max(0, len(rows)-inner.h))
	end := len(rows) - s.scroll
	start := max(0, end-inner.h)
	s.view = viewport{first: start, height: inner.h}
	clipped := &clippedScreen{Screen: a.screen, paneRect: inner}
	for y, row := range rows[start:end] {
		drawRow(clipped, y, inner.w, row, a.frame, a.glowFrame())
	}
	paintSelection(clipped, s.selection, start, inner.h)
}
