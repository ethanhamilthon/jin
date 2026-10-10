package ui

import (
	"context"
	"errors"
	"jin/internal/sources"
	"time"

	"github.com/gdamore/tcell/v3"
)

// modelsResult is a model list of one provider, fetched for one session.
type modelsResult struct {
	session  string
	provider string
	models   []string
	choices  []sources.Model
	revision string
	err      error
}

// isCycleModelKey matches Ctrl+M. Terminals without the kitty keyboard
// protocol send Enter for it, so it only works where the two are distinct.
func isCycleModelKey(ev *tcell.EventKey) bool { return isCtrl(ev, 'm', false) }

// cycleModel switches to the next in-scope model of the session's own
// provider, fetching the list once per provider.
func (a *app) cycleModel() {
	s := a.active
	if a.loadingModels {
		return
	}
	revision := sources.Revision(a.store)
	if s.modelChoices != nil && s.modelsRevision == revision {
		a.cycleChoice()
		return
	}
	if s.modelChoices == nil && s.models != nil && s.modelsFor == s.provider {
		a.switchModel()
		return
	}
	a.loadingModels = true
	session, providerID := s.id, s.provider
	go func() {
		ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
		defer cancel()
		catalog, err := sources.List(ctx, a.store, false)
		if err == nil && len(catalog.Models) == 0 && len(catalog.Errors) > 0 {
			err = errors.New(catalog.Errors[0].Message)
		}
		select {
		case a.modelsLoaded <- modelsResult{session: session, provider: providerID, choices: catalog.Models, revision: revision, err: err}:
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
	if result.choices != nil {
		s.modelChoices, s.modelsRevision = result.choices, result.revision
		a.cycleChoice()
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
