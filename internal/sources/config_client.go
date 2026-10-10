package sources

import (
	"jin/internal/provider"
	"jin/internal/store"
)

func FromConfig(db *store.DB, cfg store.Config) *provider.Client {
	for _, entry := range cfg.Providers {
		if entry.ID == cfg.ActiveProvider && entry.Config() == cfg.Provider {
			return Client(db, entry)
		}
	}
	return provider.NewClient(cfg.Provider)
}
