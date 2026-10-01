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
		'd': a.confirmDelete,
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

func (a *app) confirmDelete(name string) {
	if name == "" {
		return
	}
	options := []option{{label: "No", value: "no"}, {label: "Yes, delete", value: "yes"}}
	a.openList("Delete "+name+"?", options, "no", func(answer string) error {
		if answer == "yes" {
			if err := prompts.Delete(name); err != nil {
				return err
			}
		}
		a.showPrompts("")
		return nil
	})
}
