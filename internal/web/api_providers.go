package web

import (
	"net/http"

	"jin/internal/provider"
	"jin/internal/session"
)

type newProvider struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
	Effort  string `json:"effort"`
}

func (p newProvider) config() (provider.Config, error) {
	cfg := provider.Config{Kind: p.Kind, BaseURL: provider.NormalizeBaseURL(p.BaseURL), APIKey: p.APIKey}
	return cfg, cfg.Validate()
}

func (s *server) providerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/providers/name", api(func(r *http.Request) (any, error) {
		cfg, err := s.db.LoadConfig()
		return map[string]string{"name": session.DefaultProviderName(r.URL.Query().Get("kind"), cfg.Providers)}, err
	}))
	mux.HandleFunc("POST /api/providers/probe", api(func(r *http.Request) (any, error) {
		var body newProvider
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		cfg, err := body.config()
		if err != nil {
			return nil, err
		}
		return s.models(r.Context(), provider.NewClient(cfg), nil)
	}))
	mux.HandleFunc("POST /api/providers/probe-efforts", api(func(r *http.Request) (any, error) {
		var body newProvider
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		cfg, err := body.config()
		if err != nil {
			return nil, err
		}
		efforts, _ := provider.NewClient(cfg).Efforts(r.Context(), body.Model)
		return append([]string{}, efforts...), nil
	}))
	mux.HandleFunc("POST /api/providers", api(s.addProvider))
	mux.HandleFunc("POST /api/providers/{id}/activate", api(func(r *http.Request) (any, error) {
		var body struct{ Model, Effort string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		if err := s.db.SetActiveProvider(r.PathValue("id")); err != nil {
			return nil, err
		}
		if err := s.db.SaveModel(body.Model, body.Effort); err != nil {
			return nil, err
		}
		return done(s.m.ProvidersChanged())
	}))
	mux.HandleFunc("POST /api/providers/{id}/delete", api(func(r *http.Request) (any, error) {
		if err := s.db.DeleteProvider(r.PathValue("id")); err != nil {
			return nil, err
		}
		return done(s.m.ProvidersChanged())
	}))
	mux.HandleFunc("GET /api/providers/{id}/models", api(func(r *http.Request) (any, error) {
		client, scope, err := s.clientOf(r.PathValue("id"))
		if err != nil {
			return nil, err
		}
		if r.URL.Query().Get("all") == "1" {
			scope = nil
		}
		return s.models(r.Context(), client, scope)
	}))
	mux.HandleFunc("GET /api/providers/{id}/efforts", api(func(r *http.Request) (any, error) {
		client, _, err := s.clientOf(r.PathValue("id"))
		if err != nil {
			return nil, err
		}
		efforts, _ := client.Efforts(r.Context(), r.URL.Query().Get("model"))
		return append([]string{}, efforts...), nil
	}))
}
