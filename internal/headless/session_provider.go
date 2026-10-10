package headless

import (
	"fmt"
	"net/url"
	"strings"

	"jin/internal/store"
)

// pinProvider makes cfg use the provider named by --provider, else the one
// the session was started with, and returns the id to record. Without
// either it is the active provider. The environment still overrides it
// afterwards. A deleted session provider is an error unless JIN_BASE_URL and
// JIN_API_KEY replace it completely.
func pinProvider(cfg *store.Config, flagID string, record store.Session, getenv func(string) string) (string, error) {
	if flagID != "" {
		return useSaved(cfg, flagID)
	}
	if record.Provider == "" {
		return cfg.ActiveProvider, nil
	}
	if id, err := useSaved(cfg, record.Provider); err == nil {
		return id, nil
	}
	if strings.TrimSpace(getenv("JIN_BASE_URL")) != "" && strings.TrimSpace(getenv("JIN_API_KEY")) != "" {
		return record.Provider, nil
	}
	return "", fmt.Errorf("session provider %q no longer exists; set JIN_BASE_URL and JIN_API_KEY to override", record.Provider)
}

// useSaved switches cfg to the saved provider with this id; an unknown id
// is an error that lists the known ones.
func useSaved(cfg *store.Config, id string) (string, error) {
	var known []string
	for _, entry := range cfg.Providers {
		if entry.ID != id {
			known = append(known, entry.ID)
			continue
		}
		if entry.Disabled {
			return "", fmt.Errorf("provider is disabled: %s", entry.Name)
		}
		cfg.Provider = entry.Config()
		return entry.ID, nil
	}
	return "", fmt.Errorf("unknown provider %q (known: %s)", id, strings.Join(known, ", "))
}

// endpointOf is the base URL of the provider without credentials or query.
func endpointOf(base string) string {
	u, err := url.Parse(base)
	if err != nil {
		return ""
	}
	u.User, u.RawQuery, u.Fragment = nil, "", ""
	return u.String()
}
