package store

import (
	"slices"
	"testing"
)

func TestHooksDisabledRoundTrip(t *testing.T) {
	db, _ := openTwo(t)
	if err := db.SaveHooksDisabled([]string{"b", "a"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := db.LoadConfig()
	if err != nil || !slices.Equal(cfg.HooksDisabled, []string{"b", "a"}) {
		t.Fatalf("saved list: %v, %v", cfg.HooksDisabled, err)
	}
	if err := db.SaveHooksDisabled(nil); err != nil {
		t.Fatal(err)
	}
	if cfg, _ = db.LoadConfig(); len(cfg.HooksDisabled) != 0 {
		t.Errorf("cleared list: %v", cfg.HooksDisabled)
	}
}

func TestParseHooksDisabledIgnoresGarbage(t *testing.T) {
	if got := parseHooksDisabled("not json"); got != nil {
		t.Errorf("got %v", got)
	}
}
