package ui

import (
	"fmt"
	"math"
	"strconv"

	"jin/internal/core"
	"jin/internal/store"
)

func (s *chatSession) applyUsage(update core.Update) {
	if !update.Usage.Known {
		return
	}
	s.usage.Input += update.Usage.Input
	s.usage.Output += update.Usage.Output
	s.usage.Context = update.Usage.Input + update.Usage.Output
	if entry, ok := s.pricing.Lookup(update.Model); ok {
		s.usage.Cost += entry.Cost(update.Usage.Input, update.Usage.CachedInput, update.Usage.Output)
	}
}

func usageLine(u store.Usage) string {
	return "↑" + formatCount(u.Input) + "  ↓" + formatCount(u.Output) + "  ▭" + formatCount(u.Context) + "  $" + costAmount(u.Cost)
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
