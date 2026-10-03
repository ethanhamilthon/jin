package ui

import (
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// fillBlock paints the background of a block of rows.
func fillBlock(screen tcell.Screen, top, height, w int, style tcell.Style) {
	line := strings.Repeat(" ", max(0, w))
	for y := top; y < top+height; y++ {
		put(screen, 0, y, line, style)
	}
}

// tintBlock gives a drawn block its own background: every cell still on the
// page background takes bg, so the styles inside keep their colors.
func tintBlock(screen tcell.Screen, top, height, w int, bg color.Color) {
	for y := top; y < top+height; y++ {
		for x := 0; x < w; {
			str, style, width := screen.Get(x, y)
			if style.GetBackground() == colorBG || style.GetBackground() == color.Default {
				if str == "" {
					str = " "
				}
				screen.Put(x, y, str, style.Background(bg))
			}
			x += max(1, width)
		}
	}
}

// panelColor is the background of the panel above the input, by its source.
func (a *app) panelColor(panel *selector) color.Color {
	switch {
	case a.sel == panel:
		return colorPanel
	case a.slash != nil && a.slash.sel == panel:
		return colorSlashPanel
	case a.file != nil && a.file.sel == panel:
		return colorFilesPanel
	default:
		return colorMentionPanel
	}
}
