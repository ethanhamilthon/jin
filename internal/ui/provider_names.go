package ui

import (
	"jin/internal/provider"
	"jin/internal/store"
	"strconv"
	"strings"
)

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
