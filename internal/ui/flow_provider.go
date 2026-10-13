package ui

// openProviderFlow lists saved providers. Enter selects one, a adds one, and d deletes one.
func (a *app) openProviderFlow() { a.showProviders("") }

func (a *app) showProviders(current string) *selector {
	if current == "" {
		current = a.cfg.ActiveProvider
	}
	options := make([]option, len(a.cfg.Providers))
	for i, p := range a.cfg.Providers {
		detail := kindLabel(p.Kind) + " · " + p.BaseURL
		choices, chosen := []string{"On", "Off"}, indexOf(!p.Disabled)
		if p.Source == "cliproxy" {
			detail = "Subscription support was removed. Connect an API provider."
			choices, chosen = nil, 0
		}
		options[i] = option{label: p.Name, detail: detail, value: p.ID, choices: choices, chosen: chosen}
	}
	sel := a.openList("Providers", options, current, a.activateProvider)
	sel.twoLines = true
	sel.empty = "No providers yet · press a to add one"
	sel.hint = "←/→ on/off · Enter default · a add · d delete"
	sel.onChoice = func(id string, chosen int) error {
		entry, ok := a.configuredProvider(id)
		if !ok || entry.Source == "cliproxy" {
			return nil
		}
		if err := a.store.SetProviderEnabled(id, chosen == 0); err != nil {
			return err
		}
		return a.reloadProviders()
	}
	sel.mark = func(id string) string {
		entry, ok := a.configuredProvider(id)
		if ok && entry.Source != "cliproxy" && id == a.cfg.ActiveProvider {
			return "●"
		}
		return " "
	}
	sel.actions = map[rune]func(string){
		'a': func(string) { a.addProvider() },
		'd': func(id string) { a.confirmDeleteProvider(id) },
	}
	return sel
}

func (a *app) addProvider() { a.addAPIProvider() }

func (a *app) addAPIProvider() {
	sel := a.openList("Provider kind", onboardingKinds, "auto", func(kind string) error {
		a.addProviderOfKind(kind)
		return nil
	})
	sel.twoLines = true
}
