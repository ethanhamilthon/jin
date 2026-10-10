package ui

import (
	"context"

	"jin/internal/sources"
	"jin/internal/store"
)

// storeTitle saves the title settings and keeps them in the config.
func (a *app) storeTitle(t store.TitleSettings) error {
	if err := a.store.SaveTitle(t); err != nil {
		return err
	}
	a.cfg.Title = t
	return nil
}

// saveTitleThen saves the settings and shows the list again, with row selected.
func (a *app) saveTitleThen(t store.TitleSettings, row string) error {
	if err := a.storeTitle(t); err != nil {
		return err
	}
	a.showTitles(row)
	return nil
}

func (a *app) changeRefreshTitle(_ string, chosen int) error {
	t := a.cfg.Title
	t.Refresh = chosen == 0
	return a.storeTitle(t)
}

// pickTitleModel lists the models of all enabled providers, as /model does.
func (a *app) pickTitleModel() {
	current := sources.Key(a.cfg.Title.Provider, a.cfg.Title.Model)
	a.pickCatalogModel(current, func(model sources.Model, effort string) error {
		t := a.cfg.Title
		t.Provider, t.Model, t.Effort = model.Provider, model.ID, effort
		return a.saveTitleThen(t, "model")
	})
}

// resetTitleModel makes the session's own model name the sessions again.
func (a *app) resetTitleModel() {
	t := a.cfg.Title
	t.Provider, t.Model = "", ""
	a.report(a.saveTitleThen(t, "model"))
}

// pickTitleEffort lists the efforts of the title model, or of the default
// model when the session's model names the sessions.
func (a *app) pickTitleEffort() {
	providerID, model := a.cfg.ActiveProvider, a.cfg.Model
	if a.cfg.Title.Model != "" {
		providerID, model = a.cfg.Title.Provider, a.cfg.Title.Model
	}
	client, _, _ := sources.ClientFor(a.store, a.cfg, providerID)
	current := effortOrDefault(a.cfg.Title.Effort)
	a.openLoading("Reasoning effort · "+model, current, func(ctx context.Context) ([]option, error) {
		levels, _ := client.Efforts(ctx, model)
		return plainOptions(append([]string{defaultEffort}, levels...)), nil
	}, func(effort string) error {
		if effort == defaultEffort {
			effort = ""
		}
		t := a.cfg.Title
		t.Effort = effort
		return a.saveTitleThen(t, "effort")
	})
}
