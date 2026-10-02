package ui

// tabNames are the tabs of the panel that /sessions, /prompts and /hooks open.
var tabNames = []string{"Sessions", "Prompts", "Hooks"}

const (
	tabSessions = 0
	tabPrompts  = 1
	tabHooks    = 2
)

// tabBuilders open the panel tabs, in the order of tabNames.
func (a *app) tabBuilders() []func() *selector {
	return []func() *selector{
		a.openSessionsFlow,
		a.openPromptsFlow,
		func() *selector { return a.showHooks("") },
	}
}

// openTab shows one tab of the panel, wrapping around at both ends.
func (a *app) openTab(index int) {
	builders := a.tabBuilders()
	index = (index%len(builders) + len(builders)) % len(builders)
	sel := builders[index]()
	sel.tabbed, sel.tab = true, index
}
