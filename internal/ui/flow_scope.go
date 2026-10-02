package ui

import "context"

// openScopeFlow lists every model of the provider; Enter toggles one on or
// off and the list stays open. Models that are off leave the model picker
// and the Ctrl+M rotation.
func (a *app) openScopeFlow() {
	if !a.cfg.Provider.Ready() {
		a.openProviderFlow()
		return
	}
	client := a.client
	var sel *selector
	sel = a.openLoading("Scope models", "", func(ctx context.Context) ([]option, error) {
		models, err := client.Models(ctx)
		return plainOptions(models), err
	}, func(model string) error {
		all := make([]string, len(sel.options))
		for i, opt := range sel.options {
			all[i] = opt.value
		}
		scope := toggleScope(all, a.cfg.Scope, model)
		if err := a.store.SaveScopeFor(a.cfg.ActiveProvider, scope); err != nil {
			return err
		}
		a.cfg.Scope = scope
		return nil
	})
	sel.keepOpen = true
	sel.mark = func(model string) string {
		if a.scopeEnabled(model) {
			return "✓"
		}
		return " "
	}
}
