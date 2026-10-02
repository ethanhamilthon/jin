package ui

import "errors"

type command struct {
	label string
	run   func()
}

// tabNames are the tabs of the panel that /sessions, /prompts and /context open.
var tabNames = []string{"Sessions", "Prompts", "Context"}

const (
	tabSessions = 0
	tabPrompts  = 1
	tabContext  = 2
)

// tabBuilders open the panel tabs, in the order of tabNames.
func (a *app) tabBuilders() []func() *selector {
	return []func() *selector{
		a.openSessionsFlow,
		a.openPromptsFlow,
		func() *selector {
			return a.menuList("Context", []command{
				{"AGENTS.md files", func() { a.showAgents("") }},
				{"Hooks", func() { a.showHooks("") }},
			})
		},
	}
}

// openTab shows one tab of the panel, wrapping around at both ends.
func (a *app) openTab(index int) {
	builders := a.tabBuilders()
	index = (index%len(builders) + len(builders)) % len(builders)
	sel := builders[index]()
	sel.tabbed, sel.tab = true, index
}

func (a *app) menuList(title string, commands []command) *selector {
	options := make([]option, len(commands))
	for i, c := range commands {
		options[i] = option{label: c.label, value: c.label}
	}
	return a.openList(title, options, "", func(label string) error {
		for _, c := range commands {
			if c.label == label {
				c.run()
				return nil
			}
		}
		return errors.New("unknown command")
	})
}
