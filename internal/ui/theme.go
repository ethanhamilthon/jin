package ui

import (
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// Vercel Geist dark palette on a pure black background.
var (
	colorBG       = color.NewRGBColor(0x00, 0x00, 0x00)
	colorFG       = color.NewRGBColor(0xED, 0xED, 0xED)
	colorMuted    = color.NewRGBColor(0xA1, 0xA1, 0xA1)
	colorText     = color.NewRGBColor(0xB8, 0xB8, 0xB8)
	colorArgument = color.NewRGBColor(0x6E, 0x6E, 0x6E)
	colorDetail   = color.NewRGBColor(0x4F, 0x4F, 0x4F)
	colorDim      = color.NewRGBColor(0x66, 0x66, 0x66)
	colorBorder   = color.NewRGBColor(0x2E, 0x2E, 0x2E)
	colorRaised   = color.NewRGBColor(0x1A, 0x1A, 0x1A)
	colorBlue     = color.NewRGBColor(0x00, 0x70, 0xF3)
	colorBlueFG   = color.NewRGBColor(0x52, 0xA8, 0xFF)
	colorGreen    = color.NewRGBColor(0x62, 0xC0, 0x73)
	colorAmber    = color.NewRGBColor(0xFF, 0xB2, 0x24)
	colorRed      = color.NewRGBColor(0xFF, 0x61, 0x66)
	colorPurple   = color.NewRGBColor(0xBF, 0x7A, 0xF0)
	colorPink     = color.NewRGBColor(0xF7, 0x5F, 0x8F)
	colorTeal     = color.NewRGBColor(0x0A, 0xC7, 0xB4)
)

var (
	base       = tcell.StyleDefault.Background(colorBG).Foreground(colorFG)
	muted      = base.Foreground(colorMuted)
	dim        = base.Foreground(colorDim)
	border     = base.Foreground(colorBorder)
	accent     = base.Foreground(colorBlueFG)
	errorStyle = base.Foreground(colorRed)
	userStyle  = base.Background(colorRaised)
	insertMode = base.Foreground(colorBG).Background(colorGreen).Bold(true)
	normalMode = base.Foreground(colorFG).Background(colorBlue).Bold(true)
)

// ink fills in the black background and default foreground for styles built
// from the zero value, so no cell falls back to the terminal's own colors.
func ink(style tcell.Style) tcell.Style {
	if style.GetBackground() == color.Default {
		style = style.Background(colorBG)
	}
	if style.GetForeground() == color.Default {
		style = style.Foreground(colorFG)
	}
	return style
}

func put(screen tcell.Screen, x, y int, text string, style tcell.Style) {
	screen.PutStrStyled(x, y, text, ink(style))
}

func toolAccent(tool string) color.Color {
	switch tool {
	case "bash":
		return colorGreen
	case "read":
		return colorBlueFG
	case "write":
		return colorAmber
	case "edit":
		return colorPurple
	case "websearch":
		return colorTeal
	default:
		return colorPink
	}
}
