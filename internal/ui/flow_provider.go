package ui

import (
	"jin/internal/provider"
)

// openProviderFlow lists the saved providers: Enter makes one the active one,
// "a" adds a provider, "d" deletes one.
func (a *app) openProviderFlow() {
	a.showProviders("")
}

func (a *app) showProviders(current string) *selector {
	if current == "" {
		current = a.cfg.ActiveProvider
	}
	options := make([]option, len(a.cfg.Providers))
	for i, p := range a.cfg.Providers {
		detail := kindLabel(p.Kind) + " · " + p.BaseURL
		if p.Source == "cliproxy" {
			detail = "CLIProxyAPI · " + p.Profile
		}
		options[i] = option{label: p.Name, detail: detail, value: p.ID, choices: []string{"On", "Off"}, chosen: indexOf(!p.Disabled)}
	}
	sel := a.openList("Providers", options, current, a.activateProvider)
	sel.twoLines = true
	sel.empty = "No providers yet · press a to add one"
	sel.hint = "←/→ on/off · Enter default · a add · d delete · s sign-in"
	sel.onChoice = func(id string, chosen int) error {
		if err := a.store.SetProviderEnabled(id, chosen == 0); err != nil {
			return err
		}
		return a.reloadProviders()
	}
	sel.mark = func(id string) string {
		if id == a.cfg.ActiveProvider {
			return "●"
		}
		return " "
	}
	sel.actions = map[rune]func(string){
		'a': func(string) { a.addProvider() },
		'd': func(id string) { a.confirmDeleteProvider(id) },
		's': func(id string) { a.openSubscription(id) },
	}
	return sel
}

// addProvider asks for kind, name, URL and key, then the model, and saves the
// new provider as the active one.
func (a *app) addProvider() {
	kinds := []option{
		{label: "Subscription via CLIProxyAPI", detail: "Claude · Codex · Antigravity", value: "subscription"},
		{label: kindLabel(provider.KindOpenAI), detail: "/chat/completions", value: provider.KindOpenAI},
		{label: kindLabel(provider.KindResponses), detail: "/responses", value: provider.KindResponses},
		{label: kindLabel(provider.KindAnthropic), detail: "/v1/messages", value: provider.KindAnthropic},
	}
	sel := a.openList("Provider kind", kinds, provider.KindOpenAI, func(kind string) error {
		if kind == "subscription" {
			a.addSubscription()
		} else {
			a.addProviderOfKind(kind)
		}
		return nil
	})
	sel.twoLines = true
}
