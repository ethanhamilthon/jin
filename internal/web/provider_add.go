package web

import (
	"errors"
	"net/http"
	"strings"

	"jin/internal/session"
	"jin/internal/store"
)

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
