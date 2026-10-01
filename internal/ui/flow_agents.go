package ui

import "jin/internal/core"

// showAgents lists every AGENTS.md of the system prompt, empty ones too.
// Files can be edited but not deleted; only the project's own can be created.
func (a *app) showAgents(current string) *selector {
	files := core.AgentsFiles(a.dir)
	options := make([]option, len(files))
	hasProject := false
	for i, f := range files {
		options[i] = option{label: shortPath(f.Path), detail: string(f.Kind), value: f.Path}
		hasProject = hasProject || f.Kind == core.ContextProject
	}
	back := func() error { a.showAgents(current); return nil }
	sel := a.openList("AGENTS.md files", options, current, func(path string) error {
		return a.editFile(path, func() error { a.showAgents(path); return nil })
	})
	sel.empty = "No AGENTS.md files · press a to create one here"
	sel.hint = "Enter edit · e editor · / search"
	sel.actions = map[rune]func(string){'e': func(string) { a.chooseEditor(back) }}
	if !hasProject {
		sel.hint = "Enter edit · a create here · e editor · / search"
		sel.actions['a'] = func(string) { a.createAgents() }
	}
	return sel
}

func (a *app) createAgents() {
	path, err := core.CreateProjectAgents(a.dir)
	if err != nil {
		a.showAgents("").err = err.Error()
		return
	}
	if err := a.editFile(path, func() error { a.showAgents(path); return nil }); err != nil {
		a.showAgents(path).err = err.Error()
	}
}
