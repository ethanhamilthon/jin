package ui

const fallbackEffort = "medium"

// effortFor returns the effort the model was last used with, medium if it never was.
func (a *app) effortFor(model string) string {
	if effort, ok := a.cfg.ModelEfforts[model]; ok {
		return effort
	}
	return fallbackEffort
}

func (a *app) rememberEffort(model, effort string) {
	if a.cfg.ModelEfforts == nil {
		a.cfg.ModelEfforts = map[string]string{}
	}
	if known, ok := a.cfg.ModelEfforts[model]; ok && known == effort {
		return
	}
	a.cfg.ModelEfforts[model] = effort
	if err := a.store.SaveEfforts(a.cfg.ModelEfforts); err != nil {
		a.active.persistenceError("Model efforts were not saved", err)
	}
}
