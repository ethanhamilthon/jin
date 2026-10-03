package store

import (
	"jin/internal/pricing"
	"jin/internal/provider"
)

// Add counts one request into the totals: tokens, the context size and, when
// the model has a known price, the cost. Usage the provider did not report is
// ignored.
func (u *Usage) Add(usage provider.Usage, model string, prices pricing.Table) {
	if !usage.Known {
		return
	}
	u.Input += usage.Input
	u.Output += usage.Output
	u.Context = usage.Input + usage.Output
	if entry, ok := prices.Lookup(model); ok {
		u.Cost += entry.CostWithCacheWrite(usage.Input, usage.CachedInput, usage.CacheWriteInput, usage.Output)
	}
}
