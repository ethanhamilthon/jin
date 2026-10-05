package ui

import "path/filepath"

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
	glow, lit, running := a.paneGlow(leaf)
	f := paneFrame{label: paneTitle(leaf.session), focused: leaf == a.focused, glow: glow, lit: lit}
	if running && richColor && a.moving() {
		f.frame, f.moving = a.glowFrame(), true
	}
	f.draw(a.screen, r)
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
	inner.h = drawPaneBlocks(&clippedScreen{Screen: a.screen, paneRect: inner}, s, inner.w, inner.h, s == a.active && a.sel == nil)
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
