package pricing

import "strings"

var effortSuffixes = []string{"-high", "-medium", "-low", "-minimal", "-none"}

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
	if idx := strings.LastIndex(model, "/"); idx >= 0 {
		if entry, ok := t[model[idx+1:]]; ok {
			return entry, true
		}
	}
	for name, entry := range t {
		if strings.HasSuffix(name, "/"+model) {
			return entry, true
		}
	}
	return Entry{}, false
}

// Cost prices one request. Cached input tokens are billed at the cache-read
// rate, the rest of the input at the regular input rate.
func (e Entry) Cost(input, cached, output int) float64 {
	uncached := max(0, input-cached)
	return float64(uncached)*e.InputCostPerToken +
		float64(cached)*e.CacheReadCostPerToken +
		float64(output)*e.OutputCostPerToken
}
