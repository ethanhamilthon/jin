package web

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"jin/internal/provider"
	"jin/internal/session"
	"jin/internal/store"
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

func (s *server) addProvider(r *http.Request) (any, error) {
	var body newProvider
	if err := decode(r, &body); err != nil {
		return nil, err
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		return nil, errors.New("name is required")
	}
	cfg, err := body.config()
	if err != nil {
		return nil, err
	}
	saved, err := s.db.LoadConfig()
	if err != nil {
		return nil, err
	}
	entry := store.ProviderEntry{ID: session.NewProviderID(body.Name, saved.Providers), Name: body.Name, Kind: body.Kind, BaseURL: cfg.BaseURL, APIKey: body.APIKey}
	if err := s.db.AddProvider(entry); err != nil {
		return nil, err
	}
	if err := s.db.SetActiveProvider(entry.ID); err != nil {
		return nil, err
	}
	if err := s.db.SaveModel(body.Model, body.Effort); err != nil {
		return nil, err
	}
	return entry.ID, s.m.ProvidersChanged()
}

// clientOf builds the client of a saved provider and reads its scope.
func (s *server) clientOf(id string) (*provider.Client, []string, error) {
	cfg, err := s.db.LoadConfig()
	if err != nil {
		return nil, nil, err
	}
	for _, p := range cfg.Providers {
		if p.ID == id {
			scope, err := s.db.LoadScopeFor(id)
			return provider.NewClient(provider.Config{Kind: p.Kind, BaseURL: p.BaseURL, APIKey: p.APIKey}), scope, err
		}
	}
	return nil, nil, errors.New("no provider with id " + id)
}

type modelView struct {
	ID        string  `json:"id"`
	Context   int     `json:"context,omitempty"`
	Input     float64 `json:"input,omitempty"`
	Output    float64 `json:"output,omitempty"`
	Reasoning bool    `json:"reasoning,omitempty"`
	Vision    bool    `json:"vision,omitempty"`
}

func (s *server) models(ctx context.Context, client *provider.Client, scope []string) ([]modelView, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	ids, err := client.Models(ctx)
	if err != nil {
		return nil, err
	}
	prices := s.m.Prices()
	views := []modelView{}
	for _, id := range session.FilterScope(ids, scope) {
		view := modelView{ID: id}
		if e, ok := prices.Lookup(id); ok {
			view.Context, view.Input, view.Output = e.MaxInputTokens, e.InputCostPerToken*1e6, e.OutputCostPerToken*1e6
			view.Reasoning, view.Vision = e.ReasoningKnown && e.Reasoning, e.VisionKnown && e.Vision
		}
		views = append(views, view)
	}
	return views, nil
}
