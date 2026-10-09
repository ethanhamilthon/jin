package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3/color"

	"jin/internal/core"
	"jin/internal/tools"
)

func TestFramePathGoesAroundClockwise(t *testing.T) {
	path := framePath(paneRect{x: 2, y: 1, w: 5, h: 4})
	if len(path) != 2*(5+4)-4 {
		t.Fatalf("path has %d cells", len(path))
	}
	corners := map[int]string{0: "┌", 4: "┐", 7: "┘", 11: "└"}
	for i, glyph := range corners {
		if path[i].light != glyph {
			t.Errorf("cell %d = %q, want %q", i, path[i].light, glyph)
		}
	}
	if path[4].x != 6 || path[4].y != 1 || path[11].x != 2 || path[11].y != 4 {
		t.Fatalf("corners at %+v and %+v", path[4], path[11])
	}
}

func TestPaneGlowColors(t *testing.T) {
	a := &app{tasksRunning: map[string]int{"bg": 1, "fbg": 1}}
	pane := func(s *chatSession) *paneNode { return &paneNode{session: s} }
	cases := []struct {
		leaf    *paneNode
		focused bool
		want    color.Color
		lit     bool
		running bool
	}{
		{pane(&chatSession{ready: true}), true, colorBlueFG, true, false},
		{pane(&chatSession{ready: true, working: true}), true, colorGreen, true, true},
		{pane(&chatSession{id: "fbg", ready: true}), true, colorPurple, true, true},
		{pane(&chatSession{ready: true, working: true}), false, colorBlueFG, true, true},
		{pane(&chatSession{id: "bg", ready: true}), false, colorPurple, true, true},
		{pane(&chatSession{ready: true}), false, color.Default, false, false},
	}
	for i, c := range cases {
		a.focused = nil
		if c.focused {
			a.focused = c.leaf
		}
		got, lit, running := a.paneGlow(c.leaf)
		if got != c.want || lit != c.lit || running != c.running {
			t.Errorf("case %d: %v %v %v, want %v %v %v", i, got, lit, running, c.want, c.lit, c.running)
		}
	}
}

func TestQuestionStaysInItsPane(t *testing.T) {
	a, screen := layoutApp(t)
	left := &chatSession{id: "left", title: "left", width: 30, ready: true, history: []chatEntry{{kind: core.UpdateInfo, text: "hi"}}}
	right := &chatSession{id: "right", title: "right", width: 30, ready: true}
	right.ask = &askState{questions: []tools.Question{{Question: "Which one?", Options: []string{"A", "B"}}}}
	a.active = left
	a.panes = &paneNode{vertical: true, first: &paneNode{session: left}, second: &paneNode{session: right}}
	a.focused = a.panes.first
	a.draw()
	var text strings.Builder
	for y := range 24 {
		for x := 30; x < 60; x++ {
			glyph, _, _ := screen.Get(x, y)
			text.WriteString(glyph)
		}
		text.WriteString("\n")
	}
	for _, want := range []string{"Which one?"} {
		if !strings.Contains(text.String(), want) {
			t.Errorf("right pane lacks %q:\n%s", want, text.String())
		}
	}
}
