package ui

import (
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// paneFrame is how a pane's border looks on one frame. A lit border has a
// glow that runs around it clockwise; a still one is drawn heavy in the glow.
type paneFrame struct {
	label   string
	focused bool
	glow    color.Color
	lit     bool
	moving  bool
	frame   int
}

// paneGlow is the color around a pane: green on the focused pane, blue while
// its session works, purple while it waits on background tasks.
func (a *app) paneGlow(leaf *paneNode) (color.Color, bool) {
	s := leaf.session
	switch {
	case leaf == a.focused:
		return colorGreen, true
	case s == nil:
		return color.Default, false
	case s.working || !s.ready || (s.bash != nil && s.bash.running):
		return colorBlueFG, true
	case a.backgroundWaiting(s):
		return colorPurple, true
	}
	return color.Default, false
}

type frameCell struct {
	x, y         int
	light, heavy string
}

// framePath lists the border cells clockwise from the top left corner.
func framePath(r paneRect) []frameCell {
	right, bottom := r.x+r.w-1, r.y+r.h-1
	cells := []frameCell{{r.x, r.y, "┌", "┏"}}
	for x := r.x + 1; x < right; x++ {
		cells = append(cells, frameCell{x, r.y, "─", "━"})
	}
	cells = append(cells, frameCell{right, r.y, "┐", "┓"})
	for y := r.y + 1; y < bottom; y++ {
		cells = append(cells, frameCell{right, y, "│", "┃"})
	}
	cells = append(cells, frameCell{right, bottom, "┘", "┛"})
	for x := right - 1; x > r.x; x-- {
		cells = append(cells, frameCell{x, bottom, "─", "━"})
	}
	cells = append(cells, frameCell{r.x, bottom, "└", "┗"})
	for y := bottom - 1; y > r.y; y-- {
		cells = append(cells, frameCell{r.x, y, "│", "┃"})
	}
	return cells
}

// draw puts the border and its label on the top edge.
func (f paneFrame) draw(screen tcell.Screen, r paneRect) {
	if r.w < 2 || r.h < 2 {
		return
	}
	path := framePath(r)
	for i, c := range path {
		glyph, style := c.light, muted
		switch {
		case f.lit && f.moving:
			if sweepStrength(i, len(path), f.frame) >= 0.5 {
				glyph = c.heavy
			}
			style = base.Foreground(sweep(colorBorder, f.glow, i, len(path), f.frame))
		case f.lit:
			glyph, style = c.heavy, base.Foreground(f.glow)
		}
		put(screen, c.x, c.y, glyph, style)
	}
	labelStyle := muted
	if f.focused {
		labelStyle = paneLabel
	}
	if r.w > 6 && f.label != "" {
		put(screen, r.x+2, r.y, " "+truncate(f.label, r.w-5)+" ", labelStyle)
	}
}
