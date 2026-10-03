package ui

import (
	"slices"
	"strings"

	"jin/internal/hooks"
)

// projectPrefix marks the value of a project hook in the /hooks list.
const projectPrefix = "project:"

// showHooks lists the global hooks, then the hooks of this repository, with
// a mark for the ones that are on. Edits reach new sessions only: the system
// prompt is built when a session starts.
func (a *app) showHooks(current string) *selector {
	names, err := hooks.List()
	var options []option
	for _, name := range names {
		options = append(options, option{label: name, value: name})
	}
	project, _ := hooks.ListProject(a.dir)
	trusted := a.projectHooksTrusted()
	for _, name := range project {
		detail := "project"
		if !trusted {
			detail = "project · not trusted, t to allow"
		}
		options = append(options, option{label: name, detail: detail, value: projectPrefix + name})
	}
	sel := a.openList("Hooks", options, current, a.editHook)
	sel.empty = "No hooks yet · press a to add one"
	sel.hint = "Enter edit · a add · p add to project · d delete · t on/off · e editor · / search"
	sel.mark = func(value string) string {
		if a.hookDisabled(value) {
			return " "
		}
		return "✓"
	}
	sel.actions = map[rune]func(string){
		'a': func(string) { a.addHook(false) },
		'p': func(string) { a.addHook(true) },
		'd': func(value string) { a.confirmDelete(value, a.deleteHook, func() { a.showHooks("") }) },
		't': func(value string) { a.toggleHook(sel, value) },
		'e': func(string) { a.chooseEditor(func() error { a.showHooks(current); return nil }) },
	}
	if err != nil {
		sel.err = err.Error()
	}
	return sel
}

// hookKey is the name a hook has in the list of disabled hooks.
func (a *app) hookKey(value string) string {
	if name, ok := strings.CutPrefix(value, projectPrefix); ok {
		return hooks.ProjectKey(a.dir, name)
	}
	return value
}

func (a *app) hookDisabled(value string) bool {
	if strings.HasPrefix(value, projectPrefix) && !a.projectHooksTrusted() {
		return true
	}
	return slices.Contains(a.cfg.HooksDisabled, a.hookKey(value))
}

func (a *app) hookPath(value string) (string, error) {
	if name, ok := strings.CutPrefix(value, projectPrefix); ok {
		return hooks.ProjectPath(a.dir, name)
	}
	return hooks.Path(value)
}

func (a *app) editHook(value string) error {
	path, err := a.hookPath(value)
	if err != nil {
		return err
	}
	return a.editFile(path, func() error { a.showHooks(value); return nil })
}

func (a *app) addHook(project bool) {
	title := "New hook name"
	if project {
		title = "New project hook name (.jin/hooks)"
	}
	a.openField(title, "", false, func(name string) error {
		create, value := hooks.Create, name
		if project {
			create = func(name string) (string, error) { return hooks.CreateProject(a.dir, name) }
			value = projectPrefix + name
		}
		path, err := create(name)
		if err != nil {
			return err
		}
		return a.editFile(path, func() error { a.showHooks(value); return nil })
	})
}
