package sources

import (
	"context"
	"encoding/json"
	"errors"
	"jin/internal/cliproxy"
	"jin/internal/paths"
	"jin/internal/provider"
	"jin/internal/store"
	"strings"
)

func Client(db *store.DB, entry store.ProviderEntry) *provider.Client {
	client := provider.NewClient(entry.Config())
	if db == nil {
		return client
	}
	prefix := ""
	if entry.Source == "cliproxy" {
		prefix = cliproxy.ProfilePrefix(entry.Profile)
	}
	client.SetTarget(func(ctx context.Context, _ provider.Config, path string, payload []byte) (provider.Config, func(), error) {
		current, err := db.Provider(entry.ID)
		if err != nil {
			return provider.Config{}, nil, err
		}
		if current.Disabled {
			return provider.Config{}, nil, errors.New("provider is disabled: " + current.Name)
		}
		cfg := current.Config()
		if cfg.Kind != entry.Config().Kind || current.Source != entry.Source || current.Profile != entry.Profile {
			return cfg, nil, errors.New("provider changed; select the model again")
		}
		if current.Source != "cliproxy" {
			return cfg, nil, nil
		}
		if len(payload) > 0 {
			var body struct{ Model string }
			if json.Unmarshal(payload, &body) != nil {
				return cfg, nil, errors.New("invalid model request")
			}
			if body.Model != "" && !strings.HasPrefix(body.Model, cliproxy.ProfilePrefix(current.Profile)) {
				return cfg, nil, errors.New("model does not belong to this subscription provider")
			}
		}
		root, err := paths.Global("cliproxyapi")
		if err != nil {
			return cfg, nil, err
		}
		endpoint, key, release, err := cliproxy.Shared(root).Acquire(ctx)
		cfg.BaseURL, cfg.APIKey, cfg.Managed = endpoint, key, false
		return cfg, release, err
	}, prefix)
	return client
}

func ClientFor(db *store.DB, cfg store.Config, id string) (*provider.Client, string, bool) {
	if id == "" {
		id = cfg.ActiveProvider
	}
	for _, entry := range cfg.Providers {
		if entry.ID == id {
			return Client(db, entry), id, false
		}
	}
	if id == "" {
		return provider.NewClient(cfg.Provider), id, false
	}
	return provider.NewClient(provider.Config{}), id, true
}
