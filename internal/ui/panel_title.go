package ui

// drawRuleTitle labels the rule above a panel with its title, then shows the
// active search filter at the right end of the same row.
func (a *app) drawRuleTitle(y, w int, sel *selector) {
	rule(a.screen, y, w, sel.title)
	a.drawFilterHeader(y, w, sel, 4+len([]rune(sel.title)))
}
