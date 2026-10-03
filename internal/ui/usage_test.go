package ui

import (
	"testing"

	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
)

func TestStatusUsage(t *testing.T) {
	s := &chatSession{model: "m", usage: store.Usage{Input: 5000, Output: 700, Context: 42000, Cost: 0.5}}
	if got, want := s.statusUsage(), "↑5K  ↓700  ◫ 42K  $0.5000"; got != want {
		t.Errorf("no window, no cache: %q, want %q", got, want)
	}
	s.pricing = pricing.Table{"m": {MaxInputTokens: 200000}}
	s.cache.observe(provider.Usage{Input: 1000, CachedInput: 873, CacheKnown: true})
	if got, want := s.statusUsage(), "↑5K  ↓700  ◫ 42K/200K 21%  ↻ 87%  $0.5000"; got != want {
		t.Errorf("window and cache: %q, want %q", got, want)
	}
	s.cache.observe(provider.Usage{Input: 1000})
	if s.cache.percent != 87 {
		t.Errorf("a response without cache data must keep the last figure, got %d", s.cache.percent)
	}
}
