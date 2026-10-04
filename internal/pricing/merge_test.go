package pricing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMergePrecedence(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// LiteLLM has model-a with cost 1
	serverLite := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model-a": {"input_cost_per_token": 1.0}}`))
	}))
	defer serverLite.Close()

	// OpenRouter has model-a with cost 99 (should not overwrite) and model-b with cost 2
	serverRouter := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data": [
			{"id": "model-a", "pricing": {"prompt": "99.0"}},
			{"id": "model-b", "pricing": {"prompt": "2.0"}}
		]}`))
	}))
	defer serverRouter.Close()

	origSources := sources
	defer func() { sources = origSources }()

	sources = []source{
		{file: "test-lite.json", url: serverLite.URL, parse: parseLiteLLM},
		{file: "test-router.json", url: serverRouter.URL, parse: parseOpenRouter},
	}

	table := Load(context.Background())
	if entry, ok := table["model-a"]; !ok || entry.InputCostPerToken != 1.0 {
		t.Fatalf("model-a should keep LiteLLM cost 1.0, got %+v", entry)
	}
	if entry, ok := table["model-b"]; !ok || entry.InputCostPerToken != 2.0 {
		t.Fatalf("model-b should have OpenRouter cost 2.0, got %+v", entry)
	}
}
