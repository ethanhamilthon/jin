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

func TestInputGlowColors(t *testing.T) {
	a, _ := layoutApp(t)
	a.asyncRunning = map[string]int{}
	if _, ok := a.inputGlow(); ok {
		t.Fatal("an idle session must not glow")
	}
	a.active.working = true
	if c, _ := a.inputGlow(); c != colorBlueFG {
		t.Fatalf("working glow = %v, want blue", c)
	}
	a.active.working = false
	a.asyncRunning[a.active.id] = 1
	if c, _ := a.inputGlow(); c != colorPurple {
		t.Fatalf("background glow = %v, want purple", c)
	}
}

func TestStatusLinesSitOnBlue(t *testing.T) {
	a, screen := layoutApp(t)
	a.draw()
	for _, y := range []int{22, 23} {
		_, style, _ := screen.Get(0, y)
		if style.GetBackground() != colorBlue {
			t.Fatalf("row %d background = %v", y, style.GetBackground())
		}
	}
	_, style, _ := screen.Get(0, 21)
	if style.GetBackground() == colorBlue {
		t.Fatal("the rule above the status must stay black")
	}
	if got := rowText(screen, 20, 60); !strings.HasPrefix(got, "❯") {
		t.Fatalf("input prefix = %q", got)
	}
}
