package ui

import (
	"errors"
	"strings"

	"jin/internal/provider"
	"jin/internal/store"
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
		options[i] = option{label: p.Name, detail: kindLabel(p.Kind) + " · " + p.BaseURL, value: p.ID}
	}
	sel := a.openList("Providers", options, current, a.activateProvider)
	sel.twoLines = true
	sel.empty = "No providers yet · press a to add one"
	sel.hint = "Enter use · a add · d delete · / search"
	sel.mark = func(id string) string {
		if id == a.cfg.ActiveProvider {
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

// addProvider asks for kind, name, URL and key, then the model, and saves the
// new provider as the active one.
func (a *app) addProvider() {
	kinds := []option{
		{label: kindLabel(provider.KindOpenAI), detail: "/chat/completions", value: provider.KindOpenAI},
		{label: kindLabel(provider.KindResponses), detail: "/responses", value: provider.KindResponses},
		{label: kindLabel(provider.KindAnthropic), detail: "/v1/messages", value: provider.KindAnthropic},
	}
	sel := a.openList("Provider kind", kinds, provider.KindOpenAI, func(kind string) error {
		a.openField("Name", defaultProviderName(kind, a.cfg.Providers), false, func(name string) error {
			name = strings.TrimSpace(name)
			if name == "" {
				return errors.New("name is required")
			}
			urlHint := "https://api.openai.com/v1"
			if kind == provider.KindAnthropic {
				urlHint = "https://api.anthropic.com"
			}
			a.openField("Base URL", urlHint, false, func(baseURL string) error {
				cfg := provider.Config{Kind: kind, BaseURL: provider.NormalizeBaseURL(baseURL), APIKey: "pending"}
				if err := cfg.Validate(); err != nil {
					return err
				}
				a.openField("API key", "", true, func(key string) error {
					cfg.APIKey = key
					if err := cfg.Validate(); err != nil {
						return err
					}
					entry := store.ProviderEntry{ID: newProviderID(name, a.cfg.Providers), Name: name, Kind: kind, BaseURL: cfg.BaseURL, APIKey: key}
					a.openModelPicker(provider.NewClient(cfg), nil, func(model, effort string) error {
						if err := a.store.AddProvider(entry); err != nil {
							return err
						}
						if err := a.store.SetActiveProvider(entry.ID); err != nil {
							return err
						}
						if err := a.store.SaveModel(model, effort); err != nil {
							return err
						}
						return a.providerChanged(model, effort)
					})
					return nil
				})
				return nil
			})
			return nil
		})
		return nil
	})
	sel.twoLines = true
}
