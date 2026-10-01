package ui

import "errors"

type command struct {
	label, key string
	run        func()
}

// commands is the single list behind NORMAL-mode hotkeys and the Space menu.
func (a *app) commands() []command {
	return []command{
		{"New session", "n", a.newSession},
		{"Sessions", "s", a.openSessionsFlow},
		{"Model", "m", a.openModelFlow},
		{"Provider", "p", a.openProviderFlow},
		{"Web search", "w", a.openSearchFlow},
		{"Interrupt", "x", func() { a.active.agent.Interrupt() }},
		{"Quit", "q", a.requestQuit},
	}
}

func (a *app) runCommandKey(key string) {
	for _, c := range a.commands() {
		if c.key == key {
			c.run()
			return
		}
	}
}

func (a *app) openCommandsFlow() {
	commands := a.commands()
	options := make([]option, len(commands))
	for i, c := range commands {
		options[i] = option{label: c.label, detail: c.key, value: c.key}
	}
	a.openList("Commands", options, "", func(key string) error {
		for _, c := range commands {
			if c.key == key {
				c.run()
				return nil
			}
		}
		return errors.New("unknown command")
	})
}
