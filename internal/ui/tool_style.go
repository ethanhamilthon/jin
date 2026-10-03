package ui

import (
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// toolTint is how much of the label color goes into the row background:
// low enough to read as the page, high enough that the hue is still visible
// on a true-color terminal.
const toolTint = 0.12

// toolBackground is the page color with a faint tint of the tool's label
// color.
func toolBackground(tool string) color.Color {
	if !richColor {
		return colorBG
	}
	return mix(colorBG, toolAccent(tool), toolTint)
}

func toolRowStyle(tool string) tcell.Style {
	return base.Background(toolBackground(tool))
}
