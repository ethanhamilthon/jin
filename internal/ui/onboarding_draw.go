package ui

import (
	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

// drawOnboarding paints the first-run screen, everything centered: the
// version, the animated logo, a short guide, the three provider kinds and
// the slogan at the bottom. While a setup step is open, its panel and input
// sit at the bottom of the screen.
func (a *app) drawOnboarding(w, h int) {
	bottom := h
	if a.sel != nil {
		bottom = a.drawSetupPanel(w, h)
	}
	logoRows := len(bigLogo) + 2
	content := 1 + 1 + logoRows + 3 + len(onboardingKinds)*3
	if bottom-2 < content {
		logoRows = 0
		content -= len(bigLogo) + 2
	}
	y := max(0, (bottom-2-content)/2)
	centered(a.screen, w, y, a.version, dim)
	y += 2
	if logoRows > 0 {
		logoWidth := displaywidth.String(bigLogo[0])
		for i, line := range bigLogo {
			drawShimmer(a.screen, max(0, (w-logoWidth)/2), y+i, line, accent.Bold(true), a.glowFrame())
		}
		y += logoRows
	}
	guide, keys := "Welcome to jin. Choose the API your provider speaks.", "↑/↓ move · Enter select · s use another data folder · Ctrl+C quit"
	if a.cfg.Provider.Ready() {
		guide, keys = "One step left: choose a model and its reasoning effort.", "Enter choose the model · s use another data folder · Ctrl+C quit"
	}
	centered(a.screen, w, y, guide, bodyStyle)
	centered(a.screen, w, y+1, keys, dim)
	y += 3
	if a.cfg.Provider.Ready() {
		centered(a.screen, w, bottom-1, slogan, accent)
		return
	}
	for i, kind := range onboardingKinds {
		style, label := muted, kind.label
		if i == a.kindIndex {
			style, label = accent.Bold(true), "› "+label+" ‹"
		}
		centered(a.screen, w, y, label, style)
		centered(a.screen, w, y+1, kind.detail, dim)
		y += 3
	}
	if bottom-1 > y {
		centered(a.screen, w, bottom-1, slogan, accent)
	}
}

// drawSetupPanel draws the open setup step (a field or a list) with its
// input at the bottom and returns the first row it uses.
func (a *app) drawSetupPanel(w, h int) int {
	inputY := h - 2
	drawInput(a.screen, a.inputBox(), inputY, 1, w)
	rule(a.screen, h-1, w, "")
	rule(a.screen, inputY-1, w, "")
	height := a.selectorHeight(a.sel, h)
	if height == 0 {
		return inputY - 1
	}
	top := inputY - 1 - height
	a.drawSelector(a.sel, top, w, height)
	tintBlock(a.screen, top, height, w, colorPanel)
	a.drawRuleTitle(top-1, w, a.sel)
	return top - 1
}

func centered(screen tcell.Screen, w, y int, text string, style tcell.Style) {
	text = truncate(text, w-2)
	put(screen, max(0, (w-displaywidth.String(text))/2), y, text, style)
}
