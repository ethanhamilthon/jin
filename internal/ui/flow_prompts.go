package ui

import (
	"errors"
	"slices"

	"jin/internal/prompts"
)

func (a *app) openPromptsFlow() *selector { return a.showPrompts("") }

func (a *app) showPrompts(current string) *selector {
	infos, err := prompts.ListInfo()
	options := make([]option, len(infos))
	for i, info := range infos {
		opt := option{label: info.Name, value: info.Name}
		if info.System {
			opt.detail = "system"
		}
		options[i] = opt
	}
	sel := a.openList("Prompts", options, current, a.editPrompt)
	sel.tabbed, sel.tab = true, tabPrompts
	sel.empty = "No prompts yet · press a to add one"
	sel.hint = "Enter edit · a add · d delete · t on/off · e editor · / search · ←/→ tab"
	sel.mark = func(name string) string {
		if slices.Contains(a.cfg.PromptsDisabled, name) {
			return " "
		}
		return "✓"
	}
	sel.actions = map[rune]func(string){
		'a': func(string) { a.addPrompt() },
		'd': func(name string) {
			if prompts.IsSystem(name) {
				sel.err = "system prompt cannot be deleted"
				return
			}
			a.confirmDelete(name, a.deletePrompt, func() { a.showPrompts("") })
		},
		't': func(name string) { a.togglePrompt(sel, name) },
		'e': func(string) { a.chooseEditor(func() error { a.showPrompts(current); return nil }) },
	}
	if err != nil {
		sel.err = err.Error()
	}
	return sel
}

func (a *app) editPrompt(name string) error {
	if prompts.IsSystem(name) {
		return errors.New("system prompt cannot be edited")
	}
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

// togglePrompt switches a prompt on or off. The change reaches new sessions.
func (a *app) togglePrompt(sel *selector, name string) {
	if name == "" {
		return
	}
	disabled := slices.Clone(a.cfg.PromptsDisabled)
	if slices.Contains(disabled, name) {
		disabled = slices.DeleteFunc(disabled, func(n string) bool { return n == name })
	} else {
		disabled = append(disabled, name)
	}
	if err := a.store.SavePromptsDisabled(disabled); err != nil {
		sel.err = err.Error()
		return
	}
	a.cfg.PromptsDisabled = disabled
}

func (a *app) deletePrompt(name string) error {
	if err := prompts.Delete(name); err != nil {
		return err
	}
	disabled := slices.DeleteFunc(slices.Clone(a.cfg.PromptsDisabled), func(n string) bool { return n == name })
	if err := a.store.SavePromptsDisabled(disabled); err != nil {
		return err
	}
	a.cfg.PromptsDisabled = disabled
	return nil
}
