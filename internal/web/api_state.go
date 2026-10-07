package web

import (
	"net/http"

	"jin/internal/store"
	"jin/internal/tools"
)

type providerView struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	BaseURL string `json:"base_url"`
}

type toolView struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// configView is the configuration a page shows; API keys stay here.
type configView struct {
	Providers []providerView `json:"providers"`
	Active    string         `json:"active"`
	Model     string         `json:"model"`
	Effort    string         `json:"effort"`
	Ready     bool           `json:"ready"`
	Sound     soundView      `json:"sound"`
	Accent    string         `json:"accent"`
	Tools     []toolView     `json:"tools"`
	Scope     []string       `json:"scope"`
}

func (s *server) stateRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/state", api(func(r *http.Request) (any, error) {
		cfg, err := s.db.LoadConfig()
		if err != nil {
			return nil, err
		}
		s.mu.Lock()
		latest := s.latest
		s.mu.Unlock()
		return map[string]any{
			"version": s.version, "dir": s.dir, "latest": latest,
			"config": s.configView(cfg), "live": s.m.Live(),
		}, nil
	}))
}

func (s *server) configView(cfg store.Config) configView {
	view := configView{Active: cfg.ActiveProvider, Model: cfg.Model, Effort: cfg.Effort, Sound: soundView{cfg.Sound.Enabled, cfg.Sound.OnlyBlur, cfg.Sound.Volume}, Scope: cfg.Scope,
		Ready: cfg.Provider.Ready() && cfg.Model != "", Providers: []providerView{}}
	for _, p := range cfg.Providers {
		view.Providers = append(view.Providers, providerView{p.ID, p.Name, p.Kind, p.BaseURL})
	}
	view.Accent, _ = s.db.Setting(accentKey)
	disabled := map[string]bool{}
	for _, name := range cfg.ToolsDisabled {
		disabled[name] = true
	}
	for _, name := range tools.Catalog() {
		view.Tools = append(view.Tools, toolView{name, tools.Describe(name), !disabled[name]})
	}
	return view
}

const accentKey = "web.accent"

type soundView struct {
	Enabled  bool `json:"enabled"`
	OnlyBlur bool `json:"only_blur"`
	Volume   int  `json:"volume"`
}
