package ui

import "jin/internal/prompts"

func (a *app) openPromptsFlow() *selector { return a.showPrompts("") }

func (a *app) showPrompts(current string) *selector {
	names, err := prompts.List()
	options := make([]option, len(names))
	for i, name := range names {
		options[i] = option{label: name, value: name}
	}
	sel := a.openList("Prompts", options, current, a.editPrompt)
	sel.tabbed, sel.tab = true, tabPrompts
	sel.empty = "No prompts yet · press a to add one"
	sel.hint = "Enter edit · a add · d delete · e editor · / search · ←/→ tab"
	sel.actions = map[rune]func(string){
		'a': func(string) { a.addPrompt() },
		'd': func(name string) { a.confirmDelete(name, prompts.Delete, func() { a.showPrompts("") }) },
		'e': func(string) { a.chooseEditor(func() error { a.showPrompts(current); return nil }) },
	}
	if err != nil {
		sel.err = err.Error()
	}
	return sel
}

func (a *app) editPrompt(name string) error {
	path, err := prompts.Path(name)
	if err != nil {
		return err
	}
	return a.editFile(path, func() error { a.showPrompts(name); return nil })
}

func (a *app) addPrompt() {
	a.openField("New prompt name · folders with /", "", false, func(name string) error {
		path, err := prompts.Create(name)
		if err != nil {
			return err
		}
		return a.editFile(path, func() error { a.showPrompts(name); return nil })
	})
}
