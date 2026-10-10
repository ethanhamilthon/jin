package headless

import (
	"strings"

	"jin/internal/store"
)

// envValues are the model settings given through the environment.
type envValues struct {
	Model, Effort string
}

// applyEnv lets JIN_BASE_URL, JIN_API_KEY and JIN_PROVIDER_KIND (openai,
// responses or anthropic) replace the saved provider, each on its own, and returns JIN_MODEL and JIN_EFFORT. Nothing is saved: the
// values live for this process only.
func applyEnv(cfg *store.Config, getenv func(string) string) envValues {
	applyProviderEnv(cfg, getenv)
	return modelEnv(getenv)
}

// applyProviderEnv is the provider half of applyEnv. It is skipped when
// --provider names the provider, because the flag beats the environment.
func applyProviderEnv(cfg *store.Config, getenv func(string) string) {
	if v := strings.TrimSpace(getenv("JIN_BASE_URL")); v != "" {
		cfg.Provider.BaseURL = strings.TrimRight(v, "/")
		cfg.Provider.Managed = false
	}
	if v := strings.TrimSpace(getenv("JIN_API_KEY")); v != "" {
		cfg.Provider.APIKey = v
	}
	if v := strings.ToLower(strings.TrimSpace(getenv("JIN_PROVIDER_KIND"))); v != "" {
		cfg.Provider.Kind = v
	}
}

func modelEnv(getenv func(string) string) envValues {
	return envValues{Model: strings.TrimSpace(getenv("JIN_MODEL")), Effort: strings.TrimSpace(getenv("JIN_EFFORT"))}
}

// resolveModel picks the model: flag, then environment, then the session
// record, then the saved setting.
func resolveModel(flag string, env envValues, session store.Session, cfg store.Config) string {
	for _, v := range []string{flag, env.Model, session.Model, cfg.Model} {
		if v != "" {
			return v
		}
	}
	return ""
}

// resolveEffort follows the same order; the saved side is the effort last used
// with this model.
func resolveEffort(flag string, env envValues, session store.Session, cfg store.Config, model string) string {
	for _, v := range []string{flag, env.Effort, session.Effort} {
		if v != "" {
			return v
		}
	}
	if v, ok := cfg.ModelEfforts[store.EffortKey(cfg.ActiveProvider, model)]; ok {
		return v
	}
	if v, ok := cfg.ModelEfforts[model]; ok {
		return v
	}
	if model == cfg.Model {
		return cfg.Effort
	}
	return ""
}
