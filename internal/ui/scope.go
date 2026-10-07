package ui

import (
	"slices"

	"jin/internal/session"
)

// filterScope keeps the models inside scope. An empty scope, or one that
// matches nothing (say, left over from another provider), keeps everything.
func filterScope(models, scope []string) []string { return session.FilterScope(models, scope) }

func toggleScope(all, scope []string, model string) []string {
	return session.ToggleScope(all, scope, model)
}

// nextModel is the model after current, wrapping around; the first model when
// current is not in the list.
func nextModel(models []string, current string) string {
	if len(models) == 0 {
		return ""
	}
	if i := slices.Index(models, current); i >= 0 {
		return models[(i+1)%len(models)]
	}
	return models[0]
}

func (a *app) scopeEnabled(model string) bool {
	return len(a.cfg.Scope) == 0 || slices.Contains(a.cfg.Scope, model)
}
