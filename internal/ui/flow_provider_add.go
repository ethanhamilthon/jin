package ui

import (
	"errors"
	"jin/internal/provider"
	"jin/internal/store"
	"strings"
)

// addProviderOfKind asks for the name, URL and key of a provider of one kind,
// then its model, and saves it as the active one.
func (a *app) addProviderOfKind(kind string) {
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
}
