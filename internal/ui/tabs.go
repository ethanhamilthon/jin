package ui

// drawRuleTitle labels the rule above a panel: its title, or for the Space
// menu the row of tabs with the active one highlighted.
func (a *app) drawRuleTitle(y, w int, sel *selector) {
	if !sel.tabbed {
		rule(a.screen, y, w, sel.title)
		return
	}
	rule(a.screen, y, w, "")
	x := 2
	for i, name := range tabNames {
		label := " " + name + " "
		if x+len(label) >= w-2 {
			return
		}
		style := dim
		if i == sel.tab {
			style = accent.Bold(true)
		}
		put(a.screen, x, y, label, style)
		x += len([]rune(label)) + 1
	}
}
