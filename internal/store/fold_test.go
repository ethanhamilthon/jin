package store

import "testing"

func TestFoldRoundTripAndGarbage(t *testing.T) {
	db, _ := openTwo(t)
	if cfg, _ := db.LoadConfig(); cfg.Fold != 0 {
		t.Fatalf("default fold = %d", cfg.Fold)
	}
	if err := db.SaveFold(2); err != nil {
		t.Fatal(err)
	}
	if cfg, _ := db.LoadConfig(); cfg.Fold != 2 {
		t.Errorf("saved fold = %d", cfg.Fold)
	}
	for _, bad := range []string{"", "x", "7", "-1"} {
		if got := parseFold(bad); got != 0 {
			t.Errorf("parseFold(%q) = %d", bad, got)
		}
	}
}
