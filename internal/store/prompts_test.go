package store

import (
	"slices"
	"testing"
)

func TestPromptsDisabledRoundTrip(t *testing.T) {
	db, _ := openTwo(t)
	if err := db.SavePromptsDisabled([]string{"review", "plan"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := db.LoadConfig()
	if err != nil || !slices.Equal(cfg.PromptsDisabled, []string{"review", "plan"}) {
		t.Fatalf("saved list: %v, %v", cfg.PromptsDisabled, err)
	}
	if err := db.SavePromptsDisabled(nil); err != nil {
		t.Fatal(err)
	}
	if cfg, _ = db.LoadConfig(); len(cfg.PromptsDisabled) != 0 {
		t.Errorf("cleared list: %v", cfg.PromptsDisabled)
	}
}

func TestParsePromptsDisabledIgnoresGarbage(t *testing.T) {
	if got := parsePromptsDisabled("not json"); got != nil {
		t.Errorf("got %v", got)
	}
}
