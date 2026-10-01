package ui

import (
	"slices"

	"jin/internal/hooks"
)

// showHooks lists the hooks with a mark for the ones that are on. Edits reach
// new sessions only: the system prompt is built when a session starts.
func (a *app) showHooks(current string) *selector {
	names, err := hooks.List()
	options := make([]option, len(names))
	for i, name := range names {
		options[i] = option{label: name, value: name}
	}
	sel := a.openList("Hooks", options, current, a.editHook)
	sel.empty = "No hooks yet · press a to add one"
	sel.hint = "Enter edit · a add · d delete · t on/off · e editor · / search"
	sel.mark = func(name string) string {
		if slices.Contains(a.cfg.HooksDisabled, name) {
			return " "
		}
		return "✓"
	}
	sel.actions = map[rune]func(string){
		'a': func(string) { a.addHook() },
		'd': func(name string) { a.confirmDelete(name, a.deleteHook, func() { a.showHooks("") }) },
		't': func(name string) { a.toggleHook(sel, name) },
		'e': func(string) { a.chooseEditor(func() error { a.showHooks(current); return nil }) },
	}
	if err != nil {
		sel.err = err.Error()
	}
	return sel
}

func (a *app) editHook(name string) error {
	path, err := hooks.Path(name)
	if err != nil {
		return err
	}
	return a.editFile(path, func() error { a.showHooks(name); return nil })
}

func (a *app) addHook() {
	a.openField("New hook name", "", false, func(name string) error {
		path, err := hooks.Create(name)
		if err != nil {
			return err
		}
		return a.editFile(path, func() error { a.showHooks(name); return nil })
	})
}

func (a *app) toggleHook(sel *selector, name string) {
	if name == "" {
		return
	}
	disabled := hooks.Toggle(a.cfg.HooksDisabled, name)
	if err := a.store.SaveHooksDisabled(disabled); err != nil {
		sel.err = err.Error()
		return
	}
	a.cfg.HooksDisabled = disabled
}

func (a *app) deleteHook(name string) error {
	if err := hooks.Delete(name); err != nil {
		return err
	}
	disabled := hooks.Forget(a.cfg.HooksDisabled, name)
	if err := a.store.SaveHooksDisabled(disabled); err != nil {
		return err
	}
	a.cfg.HooksDisabled = disabled
	return nil
}
