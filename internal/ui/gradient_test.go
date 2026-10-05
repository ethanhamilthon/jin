package ui

import (
	"strings"
	"testing"
)

func TestSweepMovesLeftToRight(t *testing.T) {
	brightest := func(frame int) int {
		best, at := int32(-1), 0
		for x := range 60 {
			if _, _, b := sweep(colorBorder, colorBlueFG, x, 60, frame).RGB(); b > best {
				best, at = b, x
			}
		}
		return at
	}
	if a, b := brightest(10), brightest(12); b <= a {
		t.Fatalf("glow went from %d to %d, want rightward", a, b)
	}
}

func TestSweepStaysVisibleAndLong(t *testing.T) {
	for _, w := range []int{1, 10, 60, 120} {
		for frame := range 2 * w {
			bright, peak := 0, 0.0
			for x := range w {
				strength := sweepStrength(x, w, frame)
				peak = max(peak, strength)
				if strength >= 0.5 {
					bright++
				}
			}
			if peak != 1 || bright < max(1, 2*w/3) {
				t.Fatalf("width %d frame %d: peak %v, bright columns %d", w, frame, peak, bright)
			}
		}
	}
	if sweepStrength(0, 60, 29) != sweepStrength(56, 60, 29) {
		t.Fatal("glow must wrap symmetrically around the edges")
	}
}

func TestInputGlowThickness(t *testing.T) {
	a, screen := layoutApp(t)
	a.active.working = true
	a.drawInputRule(0, 60)
	if got := rowText(screen, 0, 60); !strings.HasPrefix(got, "━") || !strings.Contains(got, "─") {
		t.Fatalf("only the bright segment should be thick: %q", got)
	}
	a.active.working = false
	a.drawInputRule(0, 60)
	if got := rowText(screen, 0, 60); got != strings.Repeat("─", 60) {
		t.Fatalf("idle rule must stay thin: %q", got)
	}
}

func TestInputGlowColors(t *testing.T) {
	a, _ := layoutApp(t)
	a.tasksRunning = map[string]int{}
	if _, ok := a.inputGlow(); ok {
		t.Fatal("an idle session must not glow")
	}
	a.active.working = true
	if c, _ := a.inputGlow(); c != colorBlueFG {
		t.Fatalf("working glow = %v, want primary", c)
	}
	a.active.working = false
	a.tasksRunning[a.active.id] = 1
	if c, _ := a.inputGlow(); c != colorPurple {
		t.Fatalf("background glow = %v, want purple", c)
	}
	a.active.working = true
	if c, _ := a.inputGlow(); c != colorBlueFG {
		t.Fatalf("working glow = %v, want primary", c)
	}
}
