package headless

import (
	"fmt"
	"strings"

	"jin/internal/provider"
	"jin/internal/store"
)

// pinProvider makes cfg use the provider the session was started with and
// returns the id to record. Without a session it is the active provider. The
// environment still overrides it afterwards. A deleted provider is an error
// unless JIN_BASE_URL and JIN_API_KEY replace it completely.
func pinProvider(cfg *store.Config, record store.Session, getenv func(string) string) (string, error) {
	if record.Provider == "" {
		return cfg.ActiveProvider, nil
	}
	for _, entry := range cfg.Providers {
		if entry.ID != record.Provider {
			continue
		}
		kind := entry.Kind
		if kind == "" {
			kind = "openai"
		}
		cfg.Provider = provider.Config{Kind: kind, BaseURL: entry.BaseURL, APIKey: entry.APIKey}
		return entry.ID, nil
	}
	if strings.TrimSpace(getenv("JIN_BASE_URL")) != "" && strings.TrimSpace(getenv("JIN_API_KEY")) != "" {
		return record.Provider, nil
	}
	return "", fmt.Errorf("session provider %q no longer exists; set JIN_BASE_URL and JIN_API_KEY to override", record.Provider)
}
