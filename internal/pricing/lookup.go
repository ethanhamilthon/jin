package pricing

import (
	"slices"
	"strings"
)

var effortSuffixes = []string{"-xhigh", "-high", "-medium", "-low", "-minimal", "-none"}

// Lookup finds pricing for a model id: exact match first, then the suffix
// after the last "/" (provider-prefixed entries like "anthropic/claude-sonnet-5"),
// then any entry ending in "/<model>". A trailing reasoning-effort suffix
// (e.g. "-high" appended by a gateway) is stripped and retried once.
func (t Table) Lookup(model string) (Entry, bool) {
	if entry, ok := t.lookupExact(model); ok {
		return entry, true
	}
	for _, suffix := range effortSuffixes {
		if stripped, ok := strings.CutSuffix(model, suffix); ok {
			if entry, ok := t.lookupExact(stripped); ok {
				return entry, true
			}
		}
	}
	return Entry{}, false
}

func (t Table) lookupExact(model string) (Entry, bool) {
	if entry, ok := t[model]; ok {
		return entry, true
	}
	leaf := model
	if idx := strings.LastIndex(model, "/"); idx >= 0 {
		leaf = model[idx+1:]
		if entry, ok := t[leaf]; ok {
			return entry, true
		}
	}
	suffix := "/" + leaf
	var candidates []string
	for name := range t {
		if strings.HasSuffix(name, suffix) {
			candidates = append(candidates, name)
		}
	}
	if len(candidates) == 0 {
		return Entry{}, false
	}
	slices.SortFunc(candidates, func(a, b string) int {
		pa, pb := candidatePriority(a), candidatePriority(b)
		if pa != pb {
			return pa - pb
		}
		return strings.Compare(a, b)
	})
	return t[candidates[0]], true
}

// Cost prices one request. Cached input tokens are billed at the cache-read
// rate, the rest of the input at the regular input rate.
func (e Entry) Cost(input, cached, output int) float64 {
	cached = min(max(0, cached), max(0, input))
	uncached := max(0, input-cached)
	return float64(uncached)*e.InputCostPerToken +
		float64(cached)*e.CacheReadCostPerToken +
		float64(output)*e.OutputCostPerToken
}

// CostWithCacheWrite applies a separate write rate when the catalogue has one.
func (e Entry) CostWithCacheWrite(input, cached, written, output int) float64 {
	cost := e.Cost(input, cached, output)
	if e.CacheWriteCostPerToken > 0 {
		written = min(max(0, written), max(0, input-cached))
		cost += float64(written) * (e.CacheWriteCostPerToken - e.InputCostPerToken)
	}
	return cost
}
