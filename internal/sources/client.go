package sources

import (
	"context"
	"errors"

	"jin/internal/provider"
	"jin/internal/store"
)

func Client(db *store.DB, entry store.ProviderEntry) *provider.Client {
	client := provider.NewClient(entry.Config())
	if db == nil {
		return client
	}
	client.SetTarget(func(_ context.Context, _ provider.Config, _ string, _ []byte) (provider.Config, func(), error) {
		current, err := db.Provider(entry.ID)
		if err != nil {
			return provider.Config{}, nil, err
		}
		cfg := current.Config()
		if current.Source == "cliproxy" {
			return cfg, nil, errors.New("subscription support was removed. Connect an API provider.")
		}
		if current.Disabled {
			return cfg, nil, errors.New("provider is disabled: " + current.Name)
		}
		if cfg.Kind != entry.Config().Kind || current.Source != entry.Source || current.Profile != entry.Profile || current.BaseURL != entry.BaseURL {
			return cfg, nil, errors.New("provider changed; select the model again")
		}
		return cfg, nil, nil
	}, "")
	client.SetProtocolStore(
		func(model string) (string, error) { return db.LoadProtocol(entry.ID, entry.BaseURL, model) },
		func(model, protocol string) error { return db.SaveProtocol(entry.ID, entry.BaseURL, model, protocol) },
	)
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
