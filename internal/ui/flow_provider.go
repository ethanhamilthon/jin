package ui

import (
	"context"

	"jin/internal/provider"
)

const defaultEffort = "Default"

func (a *app) openProviderFlow() {
	a.openField("Base URL", a.cfg.Provider.BaseURL, false, func(baseURL string) error {
		cfg := provider.Config{BaseURL: provider.NormalizeBaseURL(baseURL), APIKey: "pending"}
		if err := cfg.Validate(); err != nil {
			return err
		}
		a.openField("API key", a.cfg.Provider.APIKey, true, func(key string) error {
			cfg.APIKey = key
			if err := cfg.Validate(); err != nil {
				return err
			}
			client := provider.NewClient(cfg)
			a.openModelPicker(client, nil, func(model, effort string) error {
				if err := a.store.SaveProvider(cfg, model, effort); err != nil {
					return err
				}
				if err := a.store.SaveScope(nil); err != nil {
					return err
				}
				a.cfg.Scope, a.modelList = nil, nil
				a.client.Configure(cfg)
				a.cfg.Provider = cfg
				a.useModel(model, effort)
				a.refreshIntro()
				return nil
			})
			return nil
		})
		return nil
	})
}

func (a *app) openModelFlow() {
	if !a.cfg.Provider.Ready() {
		a.openProviderFlow()
		return
	}
	a.openModelPicker(a.client, a.cfg.Scope, func(model, effort string) error {
		if err := a.store.SaveModel(model, effort); err != nil {
			return err
		}
		a.useModel(model, effort)
		return nil
	})
}

// openModelPicker chains the model list into the effort list for that model.
func (a *app) openModelPicker(client *provider.Client, scope []string, done func(model, effort string) error) {
	a.openLoading("Model", a.active.model, func(ctx context.Context) ([]option, error) {
		models, err := client.Models(ctx)
		return plainOptions(filterScope(models, scope)), err
	}, func(model string) error {
		current := a.active.effort
		if current == "" {
			current = defaultEffort
		}
		a.openLoading("Reasoning effort · "+model, current, func(ctx context.Context) ([]option, error) {
			efforts, _ := client.Efforts(ctx, model)
			return plainOptions(append([]string{defaultEffort}, efforts...)), nil
		}, func(effort string) error {
			if effort == defaultEffort {
				effort = ""
			}
			return done(model, effort)
		})
		return nil
	})
}

func (a *app) useModel(model, effort string) {
	a.cfg.Model, a.cfg.Effort = model, effort
	a.active.model, a.active.effort = model, effort
}

func plainOptions(values []string) []option {
	options := make([]option, len(values))
	for i, v := range values {
		options[i] = option{label: v, value: v}
	}
	return options
}
