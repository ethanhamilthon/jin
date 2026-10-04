package store

import (
	"time"

	"jin/internal/provider"
)

type Config struct {
	Provider       provider.Config
	Providers      []ProviderEntry
	ActiveProvider string
	Model          string
	Effort         string
	Editor         string
	Sound          Sound
	Voice          Voice
	Scope          []string
	Fold           int
	// Theme is the name of the color theme; Motion the animation speed.
	Theme, Motion string
	// StallTimeout is how long a stream may stay silent; 0 is the default.
	StallTimeout time.Duration

	ModelEfforts map[string]string

	HooksDisabled []string
	ToolsDisabled []string

	PromptsDisabled []string
}

const (
	keyBaseURL = "provider.base_url"
	keyAPIKey  = "provider.api_key"
	keyModel   = "model"
	keyEffort  = "effort"
	keyEditor  = "editor"
)

func (db *DB) LoadConfig() (Config, error) {
	if err := migrateProviders(db.sql); err != nil {
		return Config{}, err
	}
	values, err := db.settings()
	if err != nil {
		return Config{}, err
	}
	activeID, providers := activeProviderFrom(values)
	var activeCfg provider.Config
	if entry, ok := findProvider(providers, activeID); ok {
		kind := entry.Kind
		if kind == "" {
			kind = "openai"
		}
		activeCfg = provider.Config{Kind: kind, BaseURL: entry.BaseURL, APIKey: entry.APIKey}
	}
	scopeRaw := values[scopeKey(activeID)]
	if scopeRaw == "" && (activeID == "default" || activeID == "") {
		scopeRaw = values[keyScope]
	}

	return Config{
		Provider:        activeCfg,
		Providers:       providers,
		ActiveProvider:  activeID,
		Model:           values[keyModel],
		Effort:          values[keyEffort],
		Editor:          values[keyEditor],
		Sound:           parseSound(values),
		Voice:           parseVoice(values),
		Scope:           parseScope(scopeRaw),
		Fold:            parseFold(values[keyFold]),
		StallTimeout:    parseStallTimeout(values[keyStallTimeout]),
		Theme:           values[keyTheme],
		Motion:          parseMotion(values[keyMotion]),
		ModelEfforts:    parseEfforts(values[keyEfforts]),
		HooksDisabled:   parseHooksDisabled(values[keyHooksDisabled]),
		ToolsDisabled:   parseToolsDisabled(values[keyToolsDisabled]),
		PromptsDisabled: parsePromptsDisabled(values[keyPromptsDisabled]),
	}, nil
}
