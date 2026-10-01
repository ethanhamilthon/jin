package ui

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v3"
)

type modelsResult struct {
	models []string
	err    error
}

// isCycleModelKey matches Ctrl+M. Terminals without the kitty keyboard
// protocol send Enter for it, so it only works where the two are distinct.
func isCycleModelKey(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyCtrlM {
		return true
	}
	return ev.Key() == tcell.KeyRune && ev.Modifiers()&tcell.ModCtrl != 0 && (ev.Str() == "m" || ev.Str() == "M")
}

// cycleModel switches to the next in-scope model, fetching the list once.
func (a *app) cycleModel() {
	if !a.cfg.Provider.Ready() || a.loadingModels {
		return
	}
	if a.modelList != nil {
		a.switchModel()
		return
	}
	a.loadingModels = true
	go func() {
		ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
		defer cancel()
		models, err := a.client.Models(ctx)
		select {
		case a.modelsLoaded <- modelsResult{models: models, err: err}:
		case <-a.ctx.Done():
		}
	}()
}

func (a *app) receiveModels(result modelsResult) {
	a.loadingModels = false
	if result.err != nil {
		a.active.persistenceError("Model list was not loaded", result.err)
		return
	}
	a.modelList = result.models
	a.switchModel()
}

// switchModel restores the effort the next model was last used with.
func (a *app) switchModel() {
	next := nextModel(filterScope(a.modelList, a.cfg.Scope), a.active.model)
	if next == "" || next == a.active.model {
		return
	}
	a.rememberEffort(a.active.model, a.active.effort)
	effort := a.effortFor(next)
	if err := a.store.SaveModel(next, effort); err != nil {
		a.active.persistenceError("Model was not saved", err)
		return
	}
	a.useModel(next, effort)
}
