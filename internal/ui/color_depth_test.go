package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestNoColorUsesReverseVideo(t *testing.T) {
	noColor = true
	t.Cleanup(func() { noColor = false; applyTheme(themes[0]) })
	applyTheme(themes[0])
	if statusBar.GetAttributes()&tcell.AttrReverse == 0 {
		t.Fatal("status bar must be reversed without colors")
	}
}

func TestLowColorDropsTintAndShimmer(t *testing.T) {
	richColor = false
	t.Cleanup(func() { richColor = true })
	if toolBackground("bash") != colorBG {
		t.Fatal("tool rows must not be tinted on 256 colors")
	}
	a, screen := layoutApp(t)
	a.active.working = true
	a.frame = 3
	a.drawInputRule(0, 60)
	_, s1, _ := screen.Get(0, 0)
	_, s2, _ := screen.Get(30, 0)
	if s1.GetForeground() != s2.GetForeground() {
		t.Fatal("glow must be one plain color on 256 colors")
	}
}
