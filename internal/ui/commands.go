package ui

import "errors"

type command struct {
	label string
	run   func()
}

var tabNames = []string{"Commands", "Model & Provider", "Sessions", "Settings"}

// tabBuilders open the Space menu tabs, in the order of tabNames.
func (a *app) tabBuilders() []func() *selector {
	return []func() *selector{
		func() *selector {
			return a.menuList("Commands", []command{
				{"New session", a.newSession},
				{"Interrupt", func() { a.active.agent.Interrupt() }},
				{"Compact", a.compactSession},
				{"Handoff", a.handoffSession},
				{"Quit", a.requestQuit},
			})
		},
		func() *selector {
			return a.menuList("Model & Provider", []command{
				{"Select model", a.openModelFlow},
				{"Scope models", a.openScopeFlow},
				{"Provider", a.openProviderFlow},
			})
		},
		a.openSessionsFlow,
		func() *selector {
			return a.menuList("Settings", []command{
				{"Web search", a.openSearchFlow},
				{"Prompts", a.openPromptsFlow},
				{"Editor", func() { a.chooseEditor(func() error { return nil }) }},
			})
		},
	}
}

// openTab shows one tab of the Space menu, wrapping around at both ends.
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
