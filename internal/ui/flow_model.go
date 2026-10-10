package ui

import (
	"context"
	"jin/internal/daemon"
	"jin/internal/provider"
)

const defaultEffort = "Default"

func (a *app) openModelFlow() {
	a.openCatalogPicker()
}

// openModelPicker chains the model list into the effort list for that model.
// The effort list falls back to low, medium and high when the provider does
// not tell its levels.
func (a *app) openModelPicker(client *provider.Client, scope []string, done func(model, effort string) error) {
	table := a.pricing
	a.openLoading("Model · context · $ in / out per 1M tokens", a.active.model, func(ctx context.Context) ([]option, error) {
		models, err := client.Models(ctx)
		return modelOptions(filterScope(models, scope), table), err
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
	if a.backend != nil {
		if err := a.backend.Command(a.ctx, daemon.Command{Action: "model", Session: a.active.id, Model: model, Effort: effort}, nil); err != nil {
			a.backendError(err)
			return
		}
	}
	if a.active.provider == a.cfg.ActiveProvider {
		a.cfg.Model, a.cfg.Effort = model, effort
	}
	a.active.model, a.active.effort = model, effort
	a.rememberEffort(model, effort)
}

func plainOptions(values []string) []option {
	options := make([]option, len(values))
	for i, v := range values {
		options[i] = option{label: v, value: v}
	}
	return options
}
