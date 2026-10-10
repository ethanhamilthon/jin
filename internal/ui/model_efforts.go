package ui

import "jin/internal/store"

const fallbackEffort = "medium"

// effortFor returns the effort the model was last used with, medium if it never was.
func (a *app) effortFor(model string) string {
	if effort, ok := a.cfg.ModelEfforts[store.EffortKey(a.active.provider, model)]; ok {
		return effort
	}
	if effort, ok := a.cfg.ModelEfforts[model]; ok {
		return effort
	}
	return fallbackEffort
}

func (a *app) rememberEffort(model, effort string) {
	if a.cfg.ModelEfforts == nil {
		a.cfg.ModelEfforts = map[string]string{}
	}
	key := store.EffortKey(a.active.provider, model)
	if known, ok := a.cfg.ModelEfforts[key]; ok && known == effort {
		return
	}
	a.cfg.ModelEfforts[key] = effort
	if err := a.store.RememberEffort(a.active.provider, model, effort); err != nil {
		a.active.persistenceError("Model efforts were not saved", err)
	}
}
