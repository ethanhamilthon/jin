package ui

import (
	"jin/internal/provider"
	"jin/internal/session"
	"jin/internal/sources"
	"jin/internal/store"
)

// clientFor builds the client of a session for the provider it belongs to. An
// empty provider id (an old session) means the active one. A saved id that no
// longer exists is reported as missing and gets a client that cannot send.
func (a *app) clientFor(providerID string) (client *provider.Client, id string, missing bool) {
	return sources.ClientFor(a.store, a.cfg, providerID)
}

func defaultProviderName(kind string, existing []store.ProviderEntry) string {
	return session.DefaultProviderName(kind, existing)
}

func newProviderID(name string, existing []store.ProviderEntry) string {
	return session.NewProviderID(name, existing)
}
