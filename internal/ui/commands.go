package ui

import "errors"

type command struct {
	label string
	run   func()
}

// commands is the list behind the Space menu; NORMAL mode has no other hotkeys.
func (a *app) commands() []command {
	return []command{
		{"New session", a.newSession},
		{"Sessions", a.openSessionsFlow},
		{"Model", a.openModelFlow},
		{"Provider", a.openProviderFlow},
		{"Prompts", a.openPromptsFlow},
		{"Web search", a.openSearchFlow},
		{"Interrupt", func() { a.active.agent.Interrupt() }},
		{"Quit", a.requestQuit},
	}
}

func (a *app) openCommandsFlow() {
	commands := a.commands()
	options := make([]option, len(commands))
	for i, c := range commands {
		options[i] = option{label: c.label, value: c.label}
	}
	a.openList("Commands", options, "", func(label string) error {
		for _, c := range commands {
			if c.label == label {
				c.run()
				return nil
			}
		}
		return errors.New("unknown command")
	})
}
