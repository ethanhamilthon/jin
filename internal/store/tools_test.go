package store

import (
	"slices"
	"testing"
)

func TestToolsDisabledRoundTrip(t *testing.T) {
	db, _ := openTwo(t)
	if err := db.SaveToolsDisabled([]string{"bash", "todo"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := db.LoadConfig()
	if err != nil || !slices.Equal(cfg.ToolsDisabled, []string{"bash", "todo"}) {
		t.Fatalf("saved list: %v, %v", cfg.ToolsDisabled, err)
	}
	if err := db.SaveToolsDisabled(nil); err != nil {
		t.Fatal(err)
	}
	if cfg, _ = db.LoadConfig(); len(cfg.ToolsDisabled) != 0 {
		t.Errorf("cleared list: %v", cfg.ToolsDisabled)
	}
}

func TestParseToolsDisabledIgnoresGarbage(t *testing.T) {
	if got := parseToolsDisabled("not json"); got != nil {
		t.Errorf("got %v", got)
	}
}
