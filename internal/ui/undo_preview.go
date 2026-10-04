package ui

import "jin/internal/tools"

// previewUndo lists what /undo will restore and skip; Enter confirms, Esc cancels.
func (a *app) previewUndo(turn int, changes []tools.Change, plan tools.RevertResult) {
	var options []option
	for _, path := range plan.Restored {
		options = append(options, option{label: path, detail: "restore", value: path})
	}
	for _, path := range plan.Skipped {
		options = append(options, option{label: path, detail: "skip (changed since)", value: path})
	}
	a.openList("Undo preview · Enter confirms, Esc cancels", options, "", func(string) error {
		a.applyUndo(turn, changes)
		return nil
	})
}
