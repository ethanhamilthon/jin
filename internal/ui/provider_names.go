package ui

import (
	"jin/internal/provider"
	"jin/internal/session"
	"jin/internal/store"
)

// clientFor builds the client of a session for the provider it belongs to. An
// empty provider id (an old session) means the active one. A saved id that no
// longer exists is reported as missing and gets a client that cannot send.
func (a *app) clientFor(providerID string) (client *provider.Client, id string, missing bool) {
	if providerID == "" {
		return provider.NewClient(a.cfg.Provider), a.cfg.ActiveProvider, false
	}
	for _, p := range a.cfg.Providers {
		if p.ID == providerID {
			return provider.NewClient(provider.Config{Kind: p.Kind, BaseURL: p.BaseURL, APIKey: p.APIKey}), p.ID, false
		}
	}
	return provider.NewClient(provider.Config{}), providerID, true
}

func defaultProviderName(kind string, existing []store.ProviderEntry) string {
	return session.DefaultProviderName(kind, existing)
}

func newProviderID(name string, existing []store.ProviderEntry) string {
	return session.NewProviderID(name, existing)
}
