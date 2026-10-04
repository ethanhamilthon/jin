package ui

import (
	"slices"

	"jin/internal/tools"
)

// openToolsFlow switches tools on or off. It reaches new sessions only: the
// tool set is built when a session starts.
func (a *app) openToolsFlow() {
	names := tools.Catalog()
	options := make([]option, len(names))
	for i, name := range names {
		options[i] = option{
			label: name, detail: tools.Describe(name), value: name, choices: []string{"On", "Off"},
			chosen: indexOf(slices.Contains(a.cfg.ToolsDisabled, name)),
		}
	}
	sel := a.openList("Tools", options, "", func(string) error { return nil })
	sel.hint = "←/→ on/off · applies to new sessions · / search"
	sel.keepOpen = true
	sel.onChoice = func(name string, chosen int) error {
		var disabled []string
		for _, other := range names {
			off := slices.Contains(a.cfg.ToolsDisabled, other)
			if other == name {
				off = chosen == 1
			}
			if off {
				disabled = append(disabled, other)
			}
		}
		if err := a.store.SaveToolsDisabled(disabled); err != nil {
			return err
		}
		a.cfg.ToolsDisabled = disabled
		return nil
	}
}
