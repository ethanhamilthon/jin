package ui

import "jin/internal/daemon"

// reloadProviders re-reads the provider list and the active provider from the
// store, so the model list, scope and picker follow it.
func (a *app) reloadProviders() error {
	if a.backend != nil {
		if err := a.backend.Command(a.ctx, daemon.Command{Action: "providers-changed"}, nil); err != nil {
			return err
		}
	}
	cfg, err := a.store.LoadConfig()
	if err != nil {
		return err
	}
	a.cfg.Providers, a.cfg.ActiveProvider, a.cfg.Provider, a.cfg.Scope = cfg.Providers, cfg.ActiveProvider, cfg.Provider, cfg.Scope
	if a.client != nil {
		a.client.Configure(cfg.Provider)
	}
	a.markMissingProviders()
	return nil
}

// defaultProviderChanged updates defaults for future sessions without rebinding work.
func (a *app) defaultProviderChanged(model, effort string) error {
	if err := a.reloadProviders(); err != nil {
		return err
	}
	a.cfg.Model, a.cfg.Effort = model, effort
	a.rebindMissingSessions()
	return nil
}

// rebindMissingSessions moves the sessions whose provider was deleted to the
// provider the user just picked.
func (a *app) rebindMissingSessions() {
	for _, s := range a.sessions {
		if s.providerMissing {
			_ = a.rebindProvider(s)
		}
	}
}

// providerChanged completes provider setup and updates only an unconfigured onboarding session.
func (a *app) providerChanged(model, effort string) error {
	onboarding := a.onboarding()
	if err := a.reloadProviders(); err != nil {
		return err
	}
	a.cfg.Model, a.cfg.Effort = model, effort
	if s := a.active; onboarding && s != nil && s.provider == "" && !s.persisted && !s.working && len(s.pending) == 0 {
		s.provider = a.cfg.ActiveProvider
		client, _, _ := a.clientFor(s.provider)
		s.client.Bind(client)
		s.model, s.effort = model, effort
		if a.backend != nil {
			if err := a.chooseModelSource(s, s.provider, model, effort); err != nil {
				return err
			}
		}
		a.refreshIntro()
	}
	return nil
}
