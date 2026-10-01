package ui

import "slices"

// filterScope keeps the models inside scope. An empty scope, or one that
// matches nothing (say, left over from another provider), keeps everything.
func filterScope(models, scope []string) []string {
	if len(scope) == 0 {
		return models
	}
	var kept []string
	for _, model := range models {
		if slices.Contains(scope, model) {
			kept = append(kept, model)
		}
	}
	if len(kept) == 0 {
		return models
	}
	return kept
}

// toggleScope flips one model and returns the new scope, where nil means all
// models are on. The last enabled model cannot be switched off.
func toggleScope(all, scope []string, model string) []string {
	var enabled []string
	for _, m := range all {
		if len(scope) == 0 || slices.Contains(scope, m) {
			enabled = append(enabled, m)
		}
	}
	if slices.Contains(enabled, model) {
		if len(enabled) == 1 {
			return scope
		}
		enabled = slices.DeleteFunc(enabled, func(m string) bool { return m == model })
	} else {
		enabled = append(enabled, model)
	}
	if len(enabled) == len(all) {
		return nil
	}
	return enabled
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
