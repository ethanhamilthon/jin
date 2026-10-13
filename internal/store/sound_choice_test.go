package store

import "testing"

func TestSoundChoiceShared(t *testing.T) {
	one, two := openTwo(t)
	want := Sound{Enabled: true, OnlyBlur: true, Volume: 25, Tone: "chime"}
	if err := one.SaveSound(want); err != nil {
		t.Fatal(err)
	}
	cfg, err := two.LoadConfig()
	if err != nil || cfg.Sound != want {
		t.Fatalf("sound = %+v, err = %v", cfg.Sound, err)
	}
	want.Tone = "unknown"
	if err := one.SaveSound(want); err != nil {
		t.Fatal(err)
	}
	cfg, err = two.LoadConfig()
	if err != nil || cfg.Sound.Tone != "bell" {
		t.Fatalf("fallback = %+v, err = %v", cfg.Sound, err)
	}
}
