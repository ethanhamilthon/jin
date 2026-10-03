package ui

import "github.com/gdamore/tcell/v3"

// richColor is true on a true-color terminal. Gradients and tints need it:
// on 256 colors their in-between shades snap to a few palette entries and
// the moving glow flickers, so they are drawn as plain colors instead.
var richColor = true

// noColor is true when the terminal shows no color at all, or the user set
// NO_COLOR. Bands that are made of a background color are drawn in reverse
// video then, so they still stand out.
var noColor bool

func detectColors(screen tcell.Screen) {
	n := screen.Colors()
	richColor, noColor = n >= 1<<24, n == 0
}

// plainStyles replaces the background bands with attributes for a
// terminal without colors.
func plainStyles() {
	if !noColor {
		return
	}
	statusBar = base.Reverse(true)
	statusTitle = statusBar.Bold(true)
	statusSoft = statusBar
	statusWarn = statusBar.Bold(true).Underline(true)
	userStyle = base.Bold(true)
}
