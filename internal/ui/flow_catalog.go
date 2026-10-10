package ui

import (
	"context"
	"errors"
	"jin/internal/daemon"
	"jin/internal/sources"
	"jin/internal/store"
)

func (a *app) openCatalogPicker() {
	s := a.active
	a.pickCatalogModel(sources.Key(s.provider, s.model), func(choice sources.Model, effort string) error {
		return a.chooseModelSource(s, choice.Provider, choice.ID, effort)
	})
}

// pickCatalogModel lists the models of all enabled providers, then the
// reasoning efforts of the one chosen, and calls done with both.
func (a *app) pickCatalogModel(current string, done func(model sources.Model, effort string) error) {
	choices := map[string]sources.Model{}
	table := a.pricing
	a.openLoading("Models · all enabled providers", current, func(ctx context.Context) ([]option, error) {
		catalog, err := sources.List(ctx, a.store, false)
		if err != nil {
			return nil, err
		}
		options := []option{}
		for _, model := range catalog.Models {
			key := sources.Key(model.Provider, model.ID)
			choices[key] = model
			detail := model.ProviderName
			if facts, ok := table.Lookup(model.ID); ok {
				detail += " · " + modelFacts(facts)
			}
			options = append(options, option{label: model.ID, detail: detail, value: key})
		}
		if len(options) == 0 && len(catalog.Errors) > 0 {
			return nil, errors.New(catalog.Errors[0].Message)
		}
		return options, nil
	}, func(key string) error {
		choice, ok := choices[key]
		if !ok {
			return nil
		}
		entry, err := a.store.Provider(choice.Provider)
		if err != nil {
			return err
		}
		client := sources.Client(a.store, entry)
		a.openLoading("Reasoning effort · "+choice.ID, defaultEffort, func(ctx context.Context) ([]option, error) {
			levels, _ := client.Efforts(ctx, choice.ID)
			return plainOptions(append([]string{defaultEffort}, levels...)), nil
		}, func(effort string) error {
			if effort == defaultEffort {
				effort = ""
			}
			return done(choice, effort)
		})
		return nil
	})
}

func (a *app) chooseModelSource(s *chatSession, id, model, effort string) error {
	if a.backend != nil {
		if err := a.backend.Command(a.ctx, daemon.Command{Action: "model", Session: s.id, Provider: id, Model: model, Effort: effort}, nil); err != nil {
			return err
		}
		snap, err := a.backend.Snapshot(a.ctx, s.id)
		if err != nil {
			return err
		}
		a.backendSession(snap)
		if cfg, err := a.store.LoadConfig(); err == nil {
			a.cfg = cfg
		}
		return nil
	}
	if s != a.active {
		return nil
	}
	if s.working && id != s.provider {
		return errors.New("wait for this turn before changing providers")
	}
	entry, err := a.store.Provider(id)
	if err != nil {
		return err
	}
	onboarding := a.onboarding()
	makeDefault := onboarding || s.provider == a.cfg.ActiveProvider
	if err = a.store.ChooseModel(s.id, id, model, effort, makeDefault); err != nil {
		return err
	}
	s.client.Bind(sources.Client(a.store, entry))
	s.provider, s.providerMissing, s.model, s.effort = id, false, model, effort
	if makeDefault {
		a.cfg.ActiveProvider, a.cfg.Provider, a.cfg.Model, a.cfg.Effort = id, entry.Config(), model, effort
	}
	if a.cfg.ModelEfforts == nil {
		a.cfg.ModelEfforts = map[string]string{}
	}
	a.cfg.ModelEfforts[store.EffortKey(id, model)] = effort
	if onboarding {
		a.refreshIntro()
	}
	return nil
}
