package ui

import (
	"fmt"
	"math"
	"strconv"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/store"
)

// cacheRate is the share of the last request's input served from the
// provider's prompt cache. It stays unknown when the provider does not report
// cached tokens, so a missing figure is never shown as 0%.
type cacheRate struct {
	percent int
	known   bool
}

func (c *cacheRate) observe(usage provider.Usage) {
	if !usage.CacheKnown || usage.Input <= 0 {
		return
	}
	c.percent, c.known = min(100, usage.CachedInput*100/usage.Input), true
}

func (s *chatSession) applyUsage(update core.Update) {
	if !update.Usage.Known {
		return
	}
	s.cache.observe(update.Usage)
	s.usage.Input += update.Usage.Input
	s.usage.Output += update.Usage.Output
	s.usage.Context = update.Usage.Input + update.Usage.Output
	if entry, ok := s.pricing.Lookup(update.Model); ok {
		s.usage.Cost += entry.Cost(update.Usage.Input, update.Usage.CachedInput, update.Usage.Output)
	}
}

// applyCompacted bills the summary request and resets the context to what the
// summary itself weighs.
func (s *chatSession) applyCompacted(update core.Update) {
	s.applyUsage(update)
	if update.Usage.Known {
		s.usage.Context = update.Usage.Output
	}
}

const (
	contextIcon = "◫"
	cacheIcon   = "↻"
)

func usageLine(u store.Usage) string {
	return "↑" + formatCount(u.Input) + "  ↓" + formatCount(u.Output) + "  " + contextIcon + formatCount(u.Context) + "  $" + costAmount(u.Cost)
}

// statusUsage is the status bar variant: the context against the model's
// window when it is known, and the cache hit rate of the last request.
func (s *chatSession) statusUsage() string {
	context := formatCount(s.usage.Context)
	if window := s.window(); window > 0 {
		context += "/" + formatCount(window)
	}
	line := "↑" + formatCount(s.usage.Input) + "  ↓" + formatCount(s.usage.Output) + "  " + contextIcon + " " + context
	if s.cache.known {
		line += "  " + cacheIcon + " " + strconv.Itoa(s.cache.percent) + "%"
	}
	return line + "  $" + costAmount(s.usage.Cost)
}

func costAmount(cost float64) string {
	if cost <= 0 {
		return "—"
	}
	return fmt.Sprintf("%.4f", cost)
}

func formatCount(n int) string {
	switch {
	case n < 1000:
		return strconv.Itoa(n)
	case n < 1_000_000:
		return strconv.FormatFloat(math.Round(float64(n)/100)/10, 'f', -1, 64) + "K"
	default:
		return strconv.FormatFloat(math.Round(float64(n)/100_000)/10, 'f', -1, 64) + "M"
	}
}
