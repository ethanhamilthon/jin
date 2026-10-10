package headless

import (
	"context"
	"jin/internal/provider"
	"jin/internal/store"
	"math"
	"slices"
)

type modelEntry struct {
	ID            string   `json:"id"`
	Provider      string   `json:"provider,omitempty"`
	InputPerMTok  *float64 `json:"input_per_mtok,omitempty"`
	OutputPerMTok *float64 `json:"output_per_mtok,omitempty"`
	ContextWindow *int     `json:"context_window,omitempty"`
	Efforts       []string `json:"efforts,omitempty"`
}

// listModels reads the cache, fetching and caching once when it is empty.
// The cache belongs to the active provider, so with live the list of
// another provider is fetched and left uncached.
func listModels(ctx context.Context, db *store.DB, client *provider.Client, scope []string, all, live bool) ([]modelEntry, error) {
	var ids []string
	var err error
	if live {
		listCtx, cancel := context.WithTimeout(ctx, listTimeout)
		defer cancel()
		ids, err = client.Models(listCtx)
	} else {
		ids, err = db.LoadModelsCache()
	}
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 && !live {
		if _, err := refreshModels(ctx, db, client, false); err != nil {
			return nil, err
		}
		if ids, err = db.LoadModelsCache(); err != nil {
			return nil, err
		}
	}
	if !all && len(scope) > 0 {
		var kept []string
		for _, id := range ids {
			if slices.Contains(scope, id) {
				kept = append(kept, id)
			}
		}
		ids = kept
	}
	levels, err := db.LoadModelLevels()
	if err != nil {
		return nil, err
	}
	priceCtx, cancel := context.WithTimeout(ctx, pricingWait)
	defer cancel()
	table := loadPricing(priceCtx)
	out := make([]modelEntry, len(ids))
	for i, id := range ids {
		e := modelEntry{ID: id, Efforts: levels[id]}
		if len(table) > 0 {
			if entry, ok := table.Lookup(id); ok {
				if entry.InputCostPerToken > 0 {
					in := math.Round(entry.InputCostPerToken*1e8) / 100
					e.InputPerMTok = &in
				}
				if entry.OutputCostPerToken > 0 {
					outPrice := math.Round(entry.OutputCostPerToken*1e8) / 100
					e.OutputPerMTok = &outPrice
				}
				if entry.MaxInputTokens > 0 {
					ctxWin := entry.MaxInputTokens
					e.ContextWindow = &ctxWin
				}
			}
		}
		out[i] = e
	}
	return out, nil
}
