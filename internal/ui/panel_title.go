package ui

// drawRuleTitle labels the rule above a panel with its title.
func (a *app) drawRuleTitle(y, w int, sel *selector) {
	rule(a.screen, y, w, sel.title)
}
