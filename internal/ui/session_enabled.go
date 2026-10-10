package ui

import (
	"errors"
	"slices"

	"jin/internal/store"
)

const noEnabledText = "No enabled provider. Turn one on with /provider."

// noEnabledProvider refuses requests when providers are saved but none is on,
// or when the session's own provider is off. No request is tried then.
func (a *app) noEnabledProvider(s *chatSession) error {
	if len(a.cfg.Providers) == 0 {
		return nil
	}
	enabled := slices.ContainsFunc(a.cfg.Providers, func(p store.ProviderEntry) bool { return !p.Disabled })
	if enabled && !a.providerOff(s.provider) {
		return nil
	}
	return errors.New(noEnabledText)
}

func (a *app) providerOff(id string) bool {
	entry, ok := a.configuredProvider(id)
	return ok && entry.Disabled
}
