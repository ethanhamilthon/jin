package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"

	"jin/internal/tools"
)

func rowText(screen tcell.Screen, y, w int) string {
	var b strings.Builder
	for x := 0; x < w; x++ {
		str, _, _ := screen.Get(x, y)
		b.WriteString(str)
	}
	return strings.TrimSpace(b.String())
}

func layoutApp(t *testing.T) (*app, tcell.Screen) {
	t.Helper()
	screen, err := tcell.NewTerminfoScreenFromTty(vt.NewMockTerm(vt.MockOptSize{X: 60, Y: 24}))
	if err != nil {
		t.Fatal(err)
	}
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)
	a := &app{screen: screen, width: 60, sessions: map[string]*chatSession{}, registry: tools.NewRegistry()}
	a.active = &chatSession{width: 60, model: "m", title: "chat"}
	return a, screen
}

func TestStatusStaysAtTheBottomWhileInputAndPanelGrowUp(t *testing.T) {
	a, screen := layoutApp(t)
	a.active.input = clusters("one\ntwo\nthree")
	a.active.cursor = len(a.active.input)
	a.draw()
	if got := rowText(screen, 22, 60); !strings.Contains(got, "chat") {
		t.Errorf("title row not at h-2: %q", got)
	}
	if got := rowText(screen, 18, 60); !strings.Contains(got, "one") {
		t.Errorf("input should sit right above the bottom rule, row 18: %q", got)
	}
	if got := rowText(screen, 21, 60); !strings.HasPrefix(got, "───") {
		t.Errorf("no rule between the input and the status: %q", got)
	}
	a.sel = testSelector("alpha", "beta")
	a.draw()
	if got := rowText(screen, 22, 60); !strings.Contains(got, "chat") {
		t.Errorf("status moved when a menu opened: %q", got)
	}
	if got := rowText(screen, 20, 60); !strings.Contains(got, "Search") {
		t.Errorf("the search field should sit right above the bottom rule: %q", got)
	}
	if got := rowText(screen, 18, 60); !strings.Contains(got, "beta") {
		t.Errorf("menu options should end right above the rule: %q", got)
	}
}
