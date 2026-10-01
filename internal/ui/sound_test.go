package ui

import (
	"testing"

	"jin/internal/store"
)

func TestShouldRing(t *testing.T) {
	cases := []struct {
		sound   store.Sound
		focused bool
		want    bool
	}{
		{store.Sound{Enabled: true}, true, true},
		{store.Sound{Enabled: true}, false, true},
		{store.Sound{}, false, false},
		{store.Sound{Enabled: true, OnlyBlur: true}, true, false},
		{store.Sound{Enabled: true, OnlyBlur: true}, false, true},
	}
	for _, c := range cases {
		if got := shouldRing(c.sound, c.focused); got != c.want {
			t.Errorf("shouldRing(%+v, focused=%v) = %v", c.sound, c.focused, got)
		}
	}
}
