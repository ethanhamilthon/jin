package ui

import (
	"context"
	"errors"
	"jin/internal/sources"
	"jin/internal/store"
)

func (a *app) openCatalogPicker() {
	s := a.active
	choices := map[string]sources.Model{}
	table := a.pricing
	a.openLoading("Models · all enabled providers", sources.Key(s.provider, s.model), func(ctx context.Context) ([]option, error) {
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
		if !ok || a.active != s {
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
			return a.chooseModelSource(s, choice.Provider, choice.ID, effort)
		})
		return nil
	})
}

func (a *app) chooseModelSource(s *chatSession, id, model, effort string) error {
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
	makeDefault := s.provider == a.cfg.ActiveProvider
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
	return nil
}
