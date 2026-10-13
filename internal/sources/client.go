package sources

import (
	"context"
	"errors"

	"jin/internal/provider"
	"jin/internal/store"
)

var errRemoved = errors.New("subscription support was removed. Connect an API provider.")

func removed(entry store.ProviderEntry) bool { return entry.Source == "cliproxy" }

func Client(db *store.DB, entry store.ProviderEntry) *provider.Client {
	client := provider.NewClient(entry.Config())
	if db == nil {
		if removed(entry) {
			client.SetTarget(func(context.Context, provider.Config, string, []byte) (provider.Config, func(), error) {
				return provider.Config{}, nil, errRemoved
			}, "")
		}
		return client
	}
	client.SetTarget(func(_ context.Context, _ provider.Config, _ string, _ []byte) (provider.Config, func(), error) {
		current, err := db.Provider(entry.ID)
		if err != nil {
			return provider.Config{}, nil, err
		}
		cfg := current.Config()
		if removed(current) {
			return cfg, nil, errRemoved
		}
		if current.Disabled {
			return cfg, nil, errors.New("provider is disabled: " + current.Name)
		}
		if cfg.Kind != entry.Config().Kind || current.Source != entry.Source || current.Profile != entry.Profile || current.BaseURL != entry.BaseURL {
			return cfg, nil, errors.New("provider changed; select the model again")
		}
		return cfg, nil, nil
	}, "")
	epoch, _ := db.ProviderEpoch(entry.ID)
	client.SetProtocolStore(
		func(model string) (string, error) { return db.LoadProtocol(entry.ID, epoch, entry.BaseURL, model) },
		func(model, protocol string) error {
			return db.SaveProtocol(entry.ID, epoch, entry.BaseURL, model, protocol)
		},
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
