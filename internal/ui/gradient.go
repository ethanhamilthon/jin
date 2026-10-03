package ui

import (
	"math"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// mix blends from toward to by t in [0, 1].
func mix(from, to color.Color, t float64) color.Color {
	t = min(1, max(0, t))
	r1, g1, b1 := from.RGB()
	r2, g2, b2 := to.RGB()
	at := func(a, b int32) int32 { return a + int32(math.Round(float64(b-a)*t)) }
	return color.NewRGBColor(at(r1, r2), at(g1, g2), at(b1, b2))
}

// sweep is the color of column x of a line w wide on frame: a soft band of
// glow that runs from left to right over rest and wraps around.
func sweep(rest, glow color.Color, x, w, frame int) color.Color {
	band := max(8, w/3)
	period := w + 2*band
	head := (frame*max(2, w/30))%period - band
	distance := math.Abs(float64(x - head))
	return mix(rest, glow, 1-distance/float64(band))
}

var logoPalette []color.Color

// shimmer is the color of column x of the logo on frame: the palette flows
// across the letters from left to right.
func shimmer(x, frame int) color.Color {
	const step = 6.0
	pos := (float64(x) - float64(frame)*0.5) / step
	n := len(logoPalette)
	i := int(math.Floor(pos))
	from := logoPalette[((i%n)+n)%n]
	to := logoPalette[(((i+1)%n)+n)%n]
	return mix(from, to, pos-math.Floor(pos))
}

// inputGlow is the color that sweeps along the input's rules: blue while the
// session works in the foreground, purple while it waits on background tasks.
func (a *app) inputGlow() (color.Color, bool) {
	s := a.active
	switch {
	case s.working || !s.ready || (s.bash != nil && s.bash.running):
		return colorBlueFG, true
	case a.backgroundWaiting(s):
		return colorPurple, true
	}
	return color.Default, false
}

func (a *app) drawInputRule(y, w int) {
	glow, ok := a.inputGlow()
	if !ok {
		rule(a.screen, y, w, "")
		return
	}
	if !richColor {
		for x := range w {
			put(a.screen, x, y, "─", base.Foreground(glow))
		}
		return
	}
	for x := range w {
		put(a.screen, x, y, "─", base.Foreground(sweep(colorBorder, glow, x, w, a.frame)))
	}
}

// drawShimmer puts text one cell at a time in the logo gradient and returns
// the column after it.
func drawShimmer(screen tcell.Screen, x, y int, text string, style tcell.Style, frame int) int {
	if !richColor {
		put(screen, x, y, text, style)
		return x + displaywidth.String(text)
	}
	graphemes := displaywidth.StringGraphemes(text)
	for graphemes.Next() {
		put(screen, x, y, graphemes.Value(), style.Foreground(shimmer(x, frame)))
		x += graphemes.Width()
	}
	return x
}
