package ui

import (
	"testing"

	"jin/internal/store"
)

func TestShouldRing(t *testing.T) {
	cases := []struct {
		mode    string
		focused bool
		want    bool
	}{
		{store.SoundOn, true, true},
		{store.SoundOn, false, true},
		{store.SoundOff, false, false},
		{store.SoundBlur, true, false},
		{store.SoundBlur, false, true},
	}
	for _, c := range cases {
		if got := shouldRing(c.mode, c.focused); got != c.want {
			t.Errorf("shouldRing(%s, focused=%v) = %v", c.mode, c.focused, got)
		}
	}
}
