package headless

import (
	"testing"

	"jin/internal/provider"
	"jin/internal/store"
)

func TestApplyEnvOverridesPartly(t *testing.T) {
	cfg := store.Config{Provider: provider.Config{BaseURL: "http://saved", APIKey: "saved"}}
	env := applyEnv(&cfg, func(k string) string {
		return map[string]string{"JIN_API_KEY": " key ", "JIN_MODEL": "m", "JIN_BASE_URL": ""}[k]
	})
	if cfg.Provider.BaseURL != "http://saved" || cfg.Provider.APIKey != "key" || env.Model != "m" || env.Effort != "" {
		t.Fatalf("cfg=%+v env=%+v", cfg, env)
	}
}

func TestPrecedence(t *testing.T) {
	cfg := store.Config{Model: "saved", Effort: "low", ModelEfforts: map[string]string{"saved": "high"}}
	rec := store.Session{Model: "rec", Effort: "mid"}
	if got := resolveModel("flag", envValues{Model: "env"}, rec, cfg); got != "flag" {
		t.Error(got)
	}
	if got := resolveModel("", envValues{Model: "env"}, rec, cfg); got != "env" {
		t.Error(got)
	}
	if got := resolveModel("", envValues{}, rec, cfg); got != "rec" {
		t.Error(got)
	}
	if got := resolveModel("", envValues{}, store.Session{}, cfg); got != "saved" {
		t.Error(got)
	}
	if got := resolveEffort("", envValues{}, store.Session{}, cfg, "saved"); got != "high" {
		t.Error(got)
	}
	if got := resolveEffort("", envValues{Effort: "e"}, rec, cfg, "rec"); got != "e" {
		t.Error(got)
	}
}

func TestApplyEnvResponsesKind(t *testing.T) {
	cfg := store.Config{}
	applyEnv(&cfg, func(k string) string {
		return map[string]string{"JIN_PROVIDER_KIND": "responses"}[k]
	})
	if cfg.Provider.Kind != provider.KindResponses {
		t.Fatalf("kind = %q", cfg.Provider.Kind)
	}
}
