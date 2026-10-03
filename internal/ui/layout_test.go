package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"

	"jin/internal/provider"
	"jin/internal/store"
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
	a := &app{screen: screen, width: 60, sessions: map[string]*chatSession{}, registry: tools.NewRegistry(), cfg: readyConfig()}
	a.active = &chatSession{width: 60, model: "m", title: "chat", ready: true}
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
	if got := rowText(screen, 14, 60); !strings.Contains(got, "beta") {
		t.Errorf("the 6-row panel spans rows 13-18 above the search row, so beta is on row 14: %q", got)
	}
	if got := rowText(screen, 12, 60); !strings.HasPrefix(got, "───") {
		t.Errorf("the panel rule should sit right above its 6 rows: %q", got)
	}
}

func TestEveryPanelHasFixedHeight(t *testing.T) {
	a, _ := layoutApp(t)
	for _, count := range []int{1, 3, 20} {
		labels := make([]string, count)
		for i := range labels {
			labels[i] = "item"
		}
		sel := testSelector(labels...)
		sel.tabbed = true
		if got := a.selectorHeight(sel, 24); got != maxSelectorRows {
			t.Errorf("%d options: height %d, want %d", count, got, maxSelectorRows)
		}
	}
	for _, sel := range []*selector{testSelector("one"), {field: true}, {loading: true}, {err: "x"}} {
		if got := a.selectorHeight(sel, 24); got != maxSelectorRows {
			t.Errorf("every panel is %d rows high, got %d for %+v", maxSelectorRows, got, sel)
		}
	}
}

func TestChoosingInAMenuReturnsToInput(t *testing.T) {
	a, _ := layoutApp(t)
	ran := false
	a.openList("Menu", []option{{label: "go", value: "go"}}, "", func(string) error { ran = true; return nil })
	a.submitSelector()
	if !ran || a.sel != nil || !a.inputBox().focused {
		t.Errorf("ran=%v sel=%v, want the menu closed and input focused", ran, a.sel)
	}
}

// readyConfig has a provider and a model, so tests see the chat, not the
// first-run screen.
func readyConfig() store.Config {
	return store.Config{Provider: provider.Config{BaseURL: "http://local", APIKey: "k"}, Model: "m"}
}
