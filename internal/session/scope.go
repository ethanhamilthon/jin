package session

import "slices"

// FilterScope keeps the models in scope; an empty scope, or one that
// matches nothing, keeps them all.
func FilterScope(models, scope []string) []string {
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

// ToggleScope flips one model and returns the new scope, where nil means all
// models are on. The last enabled model cannot be switched off.
func ToggleScope(all, scope []string, model string) []string {
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
