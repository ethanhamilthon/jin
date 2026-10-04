// Package pricing resolves per-token model prices from public catalogues,
// cached under the global data directory.
package pricing

import (
	"context"
	"os"
	"time"

	"jin/internal/paths"
)

const cacheTTL = 24 * time.Hour

type Entry struct {
	InputCostPerToken      float64
	OutputCostPerToken     float64
	CacheReadCostPerToken  float64
	CacheWriteCostPerToken float64
	MaxInputTokens         int
	VisionKnown            bool
	Vision                 bool
	ReasoningKnown         bool
	Reasoning              bool
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
			if _, exists := merged[name]; !exists {
				merged[name] = entry
			}
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
		_ = writeAtomic(file, raw)
		return table
	}
	if table, err := readCache(file, src.parse); err == nil {
		return table
	}
	return Table{}
}
