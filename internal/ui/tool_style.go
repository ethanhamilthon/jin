package ui

import (
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// toolTint is how much of the label color survives in the row background:
// low enough to read as near-black, high enough that the hue is still visible
// on a true-color terminal.
const toolTint = 0.12

// toolBackground is the darkest tint of the tool's label color.
func toolBackground(tool string) color.Color {
	r, g, b := toolAccent(tool).RGB()
	return color.NewRGBColor(int32(float64(r)*toolTint), int32(float64(g)*toolTint), int32(float64(b)*toolTint))
}

func toolRowStyle(tool string) tcell.Style {
	return base.Background(toolBackground(tool))
}
