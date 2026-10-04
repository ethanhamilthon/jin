package ui

import "github.com/gdamore/tcell/v3"

func (a *app) paneFocusKey(ev *tcell.EventKey) bool {
	if ev == nil || !ev.Pressed() || ev.Modifiers()&tcell.ModAlt == 0 || a.panes == nil {
		return false
	}
	dx, dy := 0, 0
	switch ev.Key() {
	case tcell.KeyLeft:
		dx = -1
	case tcell.KeyRight:
		dx = 1
	case tcell.KeyUp:
		dy = -1
	case tcell.KeyDown:
		dy = 1
	default:
		return false
	}
	leaves := paneLeaves(a.panes)
	if len(leaves) < 2 {
		return false
	}
	w, h := a.screen.Size()
	rects := paneRects(a.panes, w, paneArea(a, w, h).h)
	moveFocus(a, leaves, rects, dx, dy)
	return true
}

func paneAbs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func moveFocus(a *app, leaves []*paneNode, rects map[*paneNode]paneRect, dx, dy int) {
	from := a.focusedLeaf()
	if from == nil {
		return
	}
	origin := rects[from]
	cx, cy := origin.x+origin.w/2, origin.y+origin.h/2
	best, bestScore := (*paneNode)(nil), int(^uint(0)>>1)
	for _, candidate := range leaves {
		if candidate == from {
			continue
		}
		r := rects[candidate]
		x, y := r.x+r.w/2, r.y+r.h/2
		primary, cross := (x-cx)*dx+(y-cy)*dy, paneAbs(x-cx)*paneAbs(dy)+paneAbs(y-cy)*paneAbs(dx)
		if primary > 0 && primary*100+cross < bestScore {
			best, bestScore = candidate, primary*100+cross
		}
	}
	if best != nil {
		a.focus(best.session)
	}
}

func (a *app) paneMouse(ev *tcell.EventMouse) bool {
	if ev == nil || a.panes == nil {
		return false
	}
	w, h := a.screen.Size()
	rects := paneRects(a.panes, w, paneArea(a, w, h).h)
	leaves := paneLeaves(a.panes)
	x, y := ev.Position()
	leaf := (*paneNode)(nil)
	for _, item := range leaves {
		if item.session != nil && item.session.selection.dragging {
			leaf = item
			break
		}
	}
	if leaf == nil {
		for _, item := range leaves {
			if rects[item].contains(x, y) {
				leaf = item
				break
			}
		}
	}
	if leaf == nil || leaf.session == nil {
		return false
	}
	if ev.Buttons()&tcell.ButtonPrimary != 0 || leaf.session.selection.dragging {
		a.focus(leaf.session)
	}
	r := rects[leaf]
	inner := paneRect{x: r.x + 1, y: r.y + 1, w: max(0, r.w-2), h: max(0, r.h-2)}
	if inner.w > 0 && inner.h > 0 && (inner.contains(x, y) || leaf.session.selection.dragging) {
		local := tcell.NewEventMouse(x-inner.x, y-inner.y, ev.Buttons(), ev.Modifiers())
		leaf.session.handleMouse(local, &clippedScreen{Screen: a.screen, paneRect: inner})
	}
	return true
}
