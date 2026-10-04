package ui

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v3"
)

// modelsResult is a model list of one provider, fetched for one session.
type modelsResult struct {
	session  string
	provider string
	models   []string
	err      error
}

// isCycleModelKey matches Ctrl+M. Terminals without the kitty keyboard
// protocol send Enter for it, so it only works where the two are distinct.
func isCycleModelKey(ev *tcell.EventKey) bool { return isCtrl(ev, 'm', false) }

// cycleModel switches to the next in-scope model of the session's own
// provider, fetching the list once per provider.
func (a *app) cycleModel() {
	s := a.active
	if !s.sessionReady() || a.loadingModels {
		return
	}
	if s.models != nil && s.modelsFor == s.provider {
		a.switchModel()
		return
	}
	a.loadingModels = true
	client, session, providerID := s.client, s.id, s.provider
	go func() {
		ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
		defer cancel()
		models, err := client.Models(ctx)
		select {
		case a.modelsLoaded <- modelsResult{session: session, provider: providerID, models: models, err: err}:
		case <-a.ctx.Done():
		}
	}()
}

// receiveModels drops a list that arrives after the focus moved to another
// session or provider.
func (a *app) receiveModels(result modelsResult) {
	a.loadingModels = false
	s := a.active
	if result.session != s.id || result.provider != s.provider {
		return
	}
	if result.err != nil {
		s.persistenceError("Model list was not loaded", result.err)
		return
	}
	s.models, s.modelsFor = result.models, result.provider
	a.switchModel()
}

// switchModel restores the effort the next model was last used with.
func (a *app) switchModel() {
	s := a.active
	next := nextModel(filterScope(s.models, a.sessionScope(s)), s.model)
	if next == "" || next == s.model {
		return
	}
	a.rememberEffort(s.model, s.effort)
	effort := a.effortFor(next)
	if s.provider == a.cfg.ActiveProvider {
		if err := a.store.SaveModel(next, effort); err != nil {
			s.persistenceError("Model was not saved", err)
			return
		}
	}
	a.useModel(next, effort)
}
