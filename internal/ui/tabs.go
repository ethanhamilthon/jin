package ui

// drawRuleTitle labels the rule above a panel: its title, or for the Esc
// menu the row of tabs with the active one highlighted.
func (a *app) drawRuleTitle(y, w int, sel *selector) {
	if !sel.tabbed {
		rule(a.screen, y, w, sel.title)
		a.drawFilterHeader(y, w, sel, 4+len([]rune(sel.title)))
		return
	}
	rule(a.screen, y, w, "")
	x, start := 2, 0
	used := 0
	for i := 0; i <= sel.tab; i++ {
		used += len(tabNames[i]) + 3
	}
	for used > w-4 && start < sel.tab {
		used -= len(tabNames[start]) + 3
		start++
	}
	for i := start; i < len(tabNames); i++ {
		name := tabNames[i]
		label := " " + name + " "
		if x+len(label) >= w-2 {
			break
		}
		style := dim
		if i == sel.tab {
			style = accent.Bold(true)
		}
		put(a.screen, x, y, label, style)
		x += len([]rune(label)) + 1
	}
	a.drawFilterHeader(y, w, sel, x)
}
