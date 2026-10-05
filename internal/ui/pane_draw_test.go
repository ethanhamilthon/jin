package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
)

func stopPaneStartup(a *app) {
	if a.active.render != nil {
		a.active.render.cancel()
	}
}

func TestMouseSelectionUsesThePaneCoordinates(t *testing.T) {
	a, screen := layoutApp(t)
	left := &chatSession{id: "left", path: "/tmp/left", title: "left", width: 60, history: []chatEntry{{kind: core.UpdateInfo, text: "other pane"}}}
	right := &chatSession{id: "right", path: "/tmp/right", title: "right", width: 60, history: []chatEntry{{kind: core.UpdateInfo, text: "target text"}}}
	a.active = left
	a.panes = &paneNode{vertical: true, first: &paneNode{session: left}, second: &paneNode{session: right}}
	a.focused = a.panes.first
	a.draw()
	if !a.paneMouse(tcell.NewEventMouse(33, 1, tcell.ButtonPrimary, tcell.ModNone)) {
		t.Fatal("pane click was not routed")
	}
	a.paneMouse(tcell.NewEventMouse(37, 1, tcell.ButtonPrimary, tcell.ModNone))
	if a.active != right || right.selection.startRow != 0 || right.selection.startCol != 2 || right.selection.endCol != 6 {
		t.Fatalf("selection routed to wrong pane or coordinates: active=%q selection=%+v", a.active.id, right.selection)
	}
	if left.selection.active {
		t.Fatal("selection changed the other pane")
	}
	a.draw()
	if glyph, _, _ := screen.Get(30, 1); glyph != "│" && glyph != "┃" {
		t.Fatalf("selection paint crossed the pane clip: %q", glyph)
	}
}

func TestClosingLastPaneRequestsProcessQuit(t *testing.T) {
	a := paneTestApp(t, 60, 24)
	a.closePane()
	if !a.quit {
		t.Fatal("closing the last pane did not request process exit")
	}
}

func TestPaneRenderSurvivesResizeAndNarrowTerminals(t *testing.T) {
	a, _ := layoutApp(t)
	left := &chatSession{id: "left", path: "/tmp/left", width: 60, rows: []chatRow{{text: "left"}}}
	right := &chatSession{id: "right", path: "/tmp/right", width: 60, rows: []chatRow{{text: "right"}}}
	a.active = left
	a.panes = &paneNode{vertical: true, first: &paneNode{session: left}, second: &paneNode{session: right}}
	a.focused = a.panes.first
	a.screen.SetSize(15, 9)
	a.width = 15
	a.draw()
	if left.width != 5 || right.width != 6 {
		t.Fatalf("pane widths were not rebuilt: left=%d right=%d", left.width, right.width)
	}
	a.screen.SetSize(8, 7)
	a.width = 8
	a.draw()
}

// TestFocusedPaneTitleSitsOnThePrimaryColor checks that the project and
// session name of the focused pane are drawn on the theme's primary color,
// and that the other pane keeps a plain label.
func TestFocusedPaneTitleSitsOnThePrimaryColor(t *testing.T) {
	a, screen := layoutApp(t)
	left := &chatSession{id: "left", path: "/tmp/left", title: "left", width: 60}
	right := &chatSession{id: "right", path: "/tmp/right", title: "right", width: 60}
	a.active = left
	a.panes = &paneNode{vertical: true, first: &paneNode{session: left}, second: &paneNode{session: right}}
	a.focused = a.panes.first
	a.draw()
	if got := rowText(screen, 0, 60); !strings.Contains(got, "left · left") || !strings.Contains(got, "right · right") {
		t.Fatalf("pane titles = %q", got)
	}
	_, focusedCell, _ := screen.Get(3, 0)
	if focusedCell.GetBackground() != colorBlue {
		t.Errorf("focused label background = %v, want the primary color %v", focusedCell.GetBackground(), colorBlue)
	}
	if focusedCell.GetForeground() != colorWhite {
		t.Errorf("focused label foreground = %v, want white", focusedCell.GetForeground())
	}
	_, unfocusedCell, _ := screen.Get(33, 0)
	if unfocusedCell.GetBackground() == colorBlue {
		t.Errorf("the unfocused pane label must stay plain, background = %v", unfocusedCell.GetBackground())
	}
	_, borderCell, _ := screen.Get(1, 0)
	if borderCell.GetBackground() == colorBlue {
		t.Errorf("the frame border must stay plain, background = %v", borderCell.GetBackground())
	}
}
