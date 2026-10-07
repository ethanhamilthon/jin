package web

import (
	"context"
	"errors"
	"time"

	"jin/internal/provider"
	"jin/internal/session"
)

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
