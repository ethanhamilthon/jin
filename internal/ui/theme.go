package ui

import (
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// The palette of the current theme. applyTheme sets every color and then
// the styles made from them.
var (
	colorBG, colorFG, colorMuted, colorText         color.Color
	colorArgument, colorDetail, colorDim            color.Color
	colorBorder, colorRaised                        color.Color
	colorBlue, colorBlueFG, colorOnBlue, colorWhite color.Color
	colorGreen, colorAmber, colorRed                color.Color
	colorPurple, colorPink, colorTeal               color.Color
	colorPanel, colorTodoPanel, colorAskPanel       color.Color
	colorSlashPanel, colorFilesPanel                color.Color
	colorMentionPanel                               color.Color
)

var (
	base, muted, dim, border, accent, errorStyle, userStyle tcell.Style
	todoPanel, askPanel, paneLabel                          tcell.Style
)

func init() { applyTheme(themes[0]) }

// styles are rebuilt from the palette whenever the theme changes.
func rebuildStyles() {
	base = tcell.StyleDefault.Background(colorBG).Foreground(colorFG)
	muted = base.Foreground(colorMuted)
	dim = base.Foreground(colorDim)
	border = base.Foreground(colorBorder)
	accent = base.Foreground(colorBlueFG)
	errorStyle = base.Foreground(colorRed)
	userStyle = base.Background(colorRaised)
	todoPanel = base.Background(colorTodoPanel)
	askPanel = base.Background(colorAskPanel)
	statusBar = base.Background(colorBlue).Foreground(colorFG)
	statusTitle = statusBar.Foreground(colorWhite).Bold(true)
	statusSoft = statusBar.Foreground(colorOnBlue)
	statusWarn = statusBar.Foreground(colorAmber).Bold(true)
	// The title of the focused pane sits on the primary color, like the
	// status bar, so the active window reads at a glance.
	paneLabel = statusBar.Foreground(colorWhite).Bold(true)
	bodyStyle = base.Foreground(colorText)
	codeStyle = base.Foreground(colorTeal)
	quoteStyle = base.Foreground(colorPurple)
	dotGreen = base.Foreground(colorGreen).Bold(true)
	dotBlue = base.Foreground(colorBlueFG).Bold(true)
	dotPurple = base.Foreground(colorPurple).Bold(true)
	logoPalette = []color.Color{colorBlueFG, colorPurple, colorPink, colorTeal}
	plainStyles()
}

// ink fills in the background and default foreground for styles built from
// the zero value, so no cell falls back to the terminal's own colors.
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
	case "read", "grep":
		return colorBlueFG
	case "write":
		return colorAmber
	case "edit":
		return colorPurple
	case "todo":
		return colorTeal
	case "ask_user":
		return colorAmber
	default:
		return colorPink
	}
}
