package ui

import (
	"testing"

	"jin/internal/store"
)

func TestMotionSpeedsAndBlur(t *testing.T) {
	a := &app{}
	for _, c := range []struct {
		motion string
		want   int
	}{{store.MotionOff, 0}, {store.MotionSlow, 5}, {store.MotionNormal, 10}, {store.MotionFast, 20}} {
		a.cfg.Motion, a.glowTenths, a.frame = c.motion, 0, 0
		for range 10 {
			a.tick()
		}
		if a.glowFrame() != c.want || a.frame != 10 {
			t.Errorf("%s: glow %d frame %d", c.motion, a.glowFrame(), a.frame)
		}
	}
	a.cfg.Motion, a.glowTenths, a.blurred = store.MotionNormal, 0, true
	a.tick()
	if a.glowTenths != 0 {
		t.Error("glow must stop while the terminal is not focused")
	}
}
