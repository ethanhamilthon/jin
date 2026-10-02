package ui

import "errors"

type command struct {
	label string
	run   func()
}

var tabNames = []string{"Commands", "Sessions", "Prompts", "Context", "Settings", "Input"}

const tabPrompts = 2

// tabBuilders open the Esc menu tabs, in the order of tabNames.
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
		a.openSessionsFlow,
		a.openPromptsFlow,
		func() *selector {
			return a.menuList("Context", []command{
				{"AGENTS.md files", func() { a.showAgents("") }},
				{"Hooks", func() { a.showHooks("") }},
			})
		},
		func() *selector {
			return a.menuList("Settings", []command{
				{"Select model", a.openModelFlow},
				{"Scope models", a.openScopeFlow},
				{"Provider", a.openProviderFlow},
				{"Sound", a.openSoundFlow},
				{"Jin docs", a.openJinDocsFlow},
				{"Editor", func() { a.chooseEditor(func() error { return nil }) }},
			})
		},
		a.openInputFlow,
	}
}

// openTab shows one tab of the Esc menu, wrapping around at both ends.
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
