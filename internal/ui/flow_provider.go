package ui

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"jin/internal/provider"
	"jin/internal/store"
)

const defaultEffort = "Default"

// kindLabel is how a provider kind reads in lists.
func kindLabel(kind string) string {
	if kind == provider.KindAnthropic {
		return "Anthropic-compatible"
	}
	return "OpenAI-compatible"
}

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
		{label: kindLabel(provider.KindAnthropic), detail: "/v1/messages", value: provider.KindAnthropic},
	}
	sel := a.openList("Provider kind", kinds, provider.KindOpenAI, func(kind string) error {
		a.openField("Name", defaultProviderName(kind, a.cfg.Providers), false, func(name string) error {
			name = strings.TrimSpace(name)
			if name == "" {
				return errors.New("name is required")
			}
			urlHint := ""
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

// activateProvider makes a saved provider the active one and asks for its model.
func (a *app) activateProvider(id string) error {
	var entry store.ProviderEntry
	found := false
	for _, p := range a.cfg.Providers {
		if p.ID == id {
			entry, found = p, true
		}
	}
	if !found {
		return errors.New("provider not found")
	}
	if id == a.cfg.ActiveProvider {
		return nil
	}
	cfg := provider.Config{Kind: entry.Kind, BaseURL: entry.BaseURL, APIKey: entry.APIKey}
	scope, err := a.store.LoadScopeFor(id)
	if err != nil {
		return err
	}
	a.openModelPicker(provider.NewClient(cfg), scope, func(model, effort string) error {
		if err := a.store.SetActiveProvider(id); err != nil {
			return err
		}
		if err := a.store.SaveModel(model, effort); err != nil {
			return err
		}
		return a.providerChanged(model, effort)
	})
	return nil
}

func (a *app) confirmDeleteProvider(id string) {
	name := id
	for _, p := range a.cfg.Providers {
		if p.ID == id {
			name = p.Name
		}
	}
	a.confirmDelete(name, func(string) error {
		if err := a.store.DeleteProvider(id); err != nil {
			return err
		}
		return a.reloadProviders()
	}, func() { a.showProviders("") })
}

// reloadProviders re-reads the provider list and the active provider from the
// store, so the model list, scope and picker follow it.
func (a *app) reloadProviders() error {
	cfg, err := a.store.LoadConfig()
	if err != nil {
		return err
	}
	a.cfg.Providers, a.cfg.ActiveProvider, a.cfg.Provider, a.cfg.Scope = cfg.Providers, cfg.ActiveProvider, cfg.Provider, cfg.Scope
	a.client.Configure(cfg.Provider)
	a.modelList = nil
	return nil
}

// providerChanged applies a new active provider with its model. A session
// that has history keeps the provider it started with, so the focused one is
// replaced by a new session; an empty one just follows.
func (a *app) providerChanged(model, effort string) error {
	if err := a.reloadProviders(); err != nil {
		return err
	}
	if s := a.active; s.persisted || s.working || len(s.pending) > 0 {
		a.newSession()
	} else {
		s.provider = a.cfg.ActiveProvider
		s.client.Configure(a.cfg.Provider)
	}
	a.useModel(model, effort)
	a.refreshIntro()
	return nil
}

// clientFor builds the client of a session for the provider it belongs to. An
// empty or unknown provider means the active one.
func (a *app) clientFor(providerID string) (*provider.Client, string) {
	for _, p := range a.cfg.Providers {
		if p.ID == providerID && providerID != "" {
			return provider.NewClient(provider.Config{Kind: p.Kind, BaseURL: p.BaseURL, APIKey: p.APIKey}), p.ID
		}
	}
	return provider.NewClient(a.cfg.Provider), a.cfg.ActiveProvider
}

func defaultProviderName(kind string, existing []store.ProviderEntry) string {
	base := "openai"
	if kind == provider.KindAnthropic {
		base = "anthropic"
	}
	return uniqueName(base, existing)
}

func uniqueName(base string, existing []store.ProviderEntry) string {
	name := base
	for n := 2; providerNameTaken(name, existing); n++ {
		name = base + "-" + strconv.Itoa(n)
	}
	return name
}

func providerNameTaken(name string, existing []store.ProviderEntry) bool {
	for _, p := range existing {
		if p.Name == name || p.ID == name {
			return true
		}
	}
	return false
}

// newProviderID turns a name into a short unique id.
func newProviderID(name string, existing []store.ProviderEntry) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "-"):
			b.WriteByte('-')
		}
	}
	id := strings.Trim(b.String(), "-")
	if id == "" {
		id = "provider"
	}
	return uniqueName(id, existing)
}

func (a *app) openModelFlow() {
	if !a.cfg.Provider.Ready() {
		a.openProviderFlow()
		return
	}
	a.openModelPicker(a.client, a.cfg.Scope, func(model, effort string) error {
		if err := a.store.SaveModel(model, effort); err != nil {
			return err
		}
		a.useModel(model, effort)
		return nil
	})
}

// openModelPicker chains the model list into the effort list for that model.
// The effort list falls back to low, medium and high when the provider does
// not tell its levels.
func (a *app) openModelPicker(client *provider.Client, scope []string, done func(model, effort string) error) {
	table := a.pricing
	a.openLoading("Model · context · $ in / out per 1M tokens", a.active.model, func(ctx context.Context) ([]option, error) {
		models, err := client.Models(ctx)
		return modelOptions(filterScope(models, scope), table), err
	}, func(model string) error {
		current := a.active.effort
		if current == "" {
			current = defaultEffort
		}
		a.openLoading("Reasoning effort · "+model, current, func(ctx context.Context) ([]option, error) {
			efforts, _ := client.Efforts(ctx, model)
			return plainOptions(append([]string{defaultEffort}, efforts...)), nil
		}, func(effort string) error {
			if effort == defaultEffort {
				effort = ""
			}
			return done(model, effort)
		})
		return nil
	})
}

func (a *app) useModel(model, effort string) {
	a.cfg.Model, a.cfg.Effort = model, effort
	a.active.model, a.active.effort = model, effort
	a.rememberEffort(model, effort)
}

func plainOptions(values []string) []option {
	options := make([]option, len(values))
	for i, v := range values {
		options[i] = option{label: v, value: v}
	}
	return options
}
