package ui

import (
	"context"
	"jin/internal/sources"
	"slices"
)

// openScopeFlow lists every model of the provider; Enter toggles one on or
// off and the list stays open. Models that are off leave the model picker
// and the Ctrl+M rotation.
func (a *app) openScopeFlow() {
	options := []option{}
	for _, p := range a.cfg.Providers {
		if !p.Disabled {
			options = append(options, option{label: p.Name, value: p.ID})
		}
	}
	if len(options) == 1 {
		if err := a.openProviderScope(options[0].value); err != nil {
			a.report(err)
		}
		return
	}
	a.openList("Provider model scope", options, a.cfg.ActiveProvider, func(id string) error { return a.openProviderScope(id) })
}

func (a *app) openProviderScope(id string) error {
	entry, err := a.store.Provider(id)
	if err != nil {
		return err
	}
	client := sources.Client(a.store, entry)
	saved, err := a.store.LoadScopeFor(id)
	if err != nil {
		return err
	}
	var sel *selector
	sel = a.openLoading("Scope models", "", func(ctx context.Context) ([]option, error) {
		models, err := client.Models(ctx)
		return plainOptions(models), err
	}, func(model string) error {
		all := make([]string, len(sel.options))
		for i, opt := range sel.options {
			all[i] = opt.value
		}
		scope := toggleScope(all, saved, model)
		if err := a.store.SaveScopeFor(id, scope); err != nil {
			return err
		}
		saved = scope
		if id == a.cfg.ActiveProvider {
			a.cfg.Scope = scope
		}
		for _, session := range a.sessions {
			session.modelChoices = nil
		}
		return nil
	})
	sel.keepOpen = true
	sel.mark = func(model string) string {
		if len(saved) == 0 || slices.Contains(saved, model) {
			return "✓"
		}
		return " "
	}
	return nil
}
