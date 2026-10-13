package startup

import (
	"jin/internal/hooks"
	"jin/internal/store"
)

// HookSettings reads the hook switches and the project trust fresh from the
// database each time the system prompt is built.
func HookSettings(db *store.DB, dir string) func() hooks.Settings {
	if db == nil {
		return nil
	}
	return func() hooks.Settings {
		cfg, _ := db.LoadConfig()
		trust, _ := db.HooksTrust(dir)
		return hooks.Settings{Disabled: cfg.HooksDisabled, Trusted: trust == store.Trusted}
	}
}
