package ui

import (
	"slices"
	"strings"

	"jin/internal/hooks"
)

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
