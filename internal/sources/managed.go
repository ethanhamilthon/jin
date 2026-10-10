package sources

import (
	"context"
	"errors"
	"jin/internal/cliproxy"
	"jin/internal/paths"
	"jin/internal/store"
)

func Managed() (*cliproxy.Client, string, error) {
	root, err := paths.Global("cliproxyapi")
	if err != nil {
		return nil, "", err
	}
	return cliproxy.Shared(root), root, nil
}

func Enable(ctx context.Context, db *store.DB, id string, enabled bool) error {
	p, err := db.Provider(id)
	if err != nil {
		return err
	}
	if err = db.SetProviderEnabled(id, enabled); err != nil {
		return err
	}
	if p.Source == "cliproxy" {
		client, root, err := Managed()
		if err != nil {
			return err
		}
		if cliproxy.Installed(root).Version != "" {
			return client.Sync(ctx, p.Profile, !enabled)
		}
	}
	return nil
}

func AddManaged(db *store.DB, profile string) error {
	switch profile {
	case "claude", "codex", "antigravity":
	default:
		return errors.New("unsupported subscription provider")
	}
	list, _, err := db.LoadProviders()
	if err != nil {
		return err
	}
	for _, p := range list {
		if p.Source == "cliproxy" && p.Profile == profile {
			return nil
		}
	}
	names := map[string]string{"claude": "Claude subscription", "codex": "Codex subscription", "antigravity": "Antigravity subscription"}
	return db.AddProvider(store.ProviderEntry{ID: "subscription-" + profile, Name: names[profile], Kind: "openai", Source: "cliproxy", Profile: profile})
}
