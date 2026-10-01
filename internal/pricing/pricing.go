// Package pricing resolves per-token model prices from public catalogues,
// cached under the global data directory.
package pricing

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"jin/internal/paths"
)

const cacheTTL = 24 * time.Hour

type Entry struct {
	InputCostPerToken     float64
	OutputCostPerToken    float64
	CacheReadCostPerToken float64
	MaxInputTokens        int
}

type Table map[string]Entry

type source struct {
	file  string
	url   string
	parse func([]byte) (Table, error)
}

// LiteLLM covers a wide catalogue of established models; OpenRouter tracks
// new releases faster and is merged on top for models LiteLLM hasn't listed yet.
var sources = []source{
	{file: "pricing-litellm.json", url: "https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json", parse: parseLiteLLM},
	{file: "pricing-openrouter.json", url: "https://openrouter.ai/api/v1/models", parse: parseOpenRouter},
}

// Load returns the merged pricing table. Per source it prefers a fresh cache,
// then a network fetch, then a stale cache, then nothing.
func Load(ctx context.Context) Table {
	merged := make(Table)
	for _, src := range sources {
		for name, entry := range loadSource(ctx, src) {
			merged[name] = entry
		}
	}
	return merged
}

func loadSource(ctx context.Context, src source) Table {
	file, err := paths.Global(src.file)
	if err != nil {
		return Table{}
	}
	if info, err := os.Stat(file); err == nil && time.Since(info.ModTime()) < cacheTTL {
		if table, err := readCache(file, src.parse); err == nil {
			return table
		}
	}
	if table, raw, err := fetch(ctx, src.url, src.parse); err == nil {
		_ = os.MkdirAll(filepath.Dir(file), 0o700)
		_ = os.WriteFile(file, raw, 0o600)
		return table
	}
	if table, err := readCache(file, src.parse); err == nil {
		return table
	}
	return Table{}
}

func readCache(file string, parse func([]byte) (Table, error)) (Table, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return parse(data)
}

func fetch(ctx context.Context, url string, parse func([]byte) (Table, error)) (Table, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, nil, err
	}
	table, err := parse(data)
	if err != nil {
		return nil, nil, err
	}
	return table, data, nil
}
