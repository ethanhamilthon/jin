package ui

import (
	"jin/internal/sounds"
	"jin/internal/store"
	"testing"
)

func TestSoundFlowHasSixteenSharedTones(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	a, _ := layoutApp(t)
	a.store = db
	a.cfg.Sound = store.Sound{Enabled: true, Volume: 0, Tone: "pluck"}
	a.openSoundFlow()
	var found bool
	for _, option := range a.sel.options {
		if option.value == "tone" {
			found = true
			if len(option.choices) != 16 || option.choices[option.chosen] != "Pluck" {
				t.Fatalf("tone choices = %+v", option)
			}
		}
	}
	if !found {
		t.Fatal("tone row is missing")
	}
	if err := a.changeSound("tone", 2); err != nil {
		t.Fatal(err)
	}
	cfg, err := db.LoadConfig()
	if err != nil || cfg.Sound.Tone != sounds.Catalog()[2].ID || a.cfg.Sound.Tone != cfg.Sound.Tone {
		t.Fatalf("saved sound = %+v, err = %v", cfg.Sound, err)
	}
}
