package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func paneTestApp(t *testing.T, width, height int) *app {
	t.Helper()
	a := startingApp(t)
	a.screen.SetSize(width, height)
	a.width = width
	root := &chatSession{id: "root", path: a.dir, width: width, ready: true, input: clusters("root draft")}
	a.sessions[root.id], a.active = root, root
	a.initPanes()
	t.Cleanup(func() {
		for _, s := range a.sessions {
			if s.updatesOut != nil {
				close(s.updatesOut)
			}
			if s.render != nil {
				s.render.cancel()
			}
			if s.stop != nil {
				s.stop()
			}
		}
	})
	return a
}

func TestNestedPaneSplitsFocusDraftsAndLimit(t *testing.T) {
	a := paneTestApp(t, 60, 24)
	root := a.active
	if !a.splitPane(true) {
		t.Fatal("first vertical split was rejected")
	}
	stopPaneStartup(a)
	right := a.active
	right.input = clusters("right draft")
	if !a.splitPane(false) {
		t.Fatal("nested horizontal split was rejected")
	}
	stopPaneStartup(a)
	if !a.paneFocusKey(tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModAlt)) {
		t.Fatal("Alt+Left was not consumed")
	}
	if a.active != root {
		t.Fatal("Alt+Left did not focus the left pane")
	}
	if !a.splitPane(false) || len(paneLeaves(a.panes)) != maxPanes {
		t.Fatal("four-pane nested layout was not created")
	}
	stopPaneStartup(a)
	before := a.panes
	if a.splitPane(true) || a.panes != before {
		t.Fatal("pane limit changed the layout")
	}
	a.focus(right)
	a.focus(root)
	if strings.Join(root.input, "") != "root draft" || strings.Join(right.input, "") != "right draft" {
		t.Fatal("focusing panes lost a draft")
	}
	if len(paneLeaves(a.panes)) != maxPanes {
		t.Fatal("focusing a visible session duplicated or removed a pane")
	}
	unseen := &chatSession{id: "unseen", path: a.dir, input: clusters("unseen draft")}
	a.sessions[unseen.id] = unseen
	a.focus(unseen)
	if a.paneForSession(unseen) == nil || a.sessions[root.id] != root || strings.Join(root.input, "") != "root draft" {
		t.Fatal("opening another session did not replace only the focused pane")
	}
}

func TestClosePaneLeavesItsRunningSessionAlive(t *testing.T) {
	a := paneTestApp(t, 60, 24)
	if !a.splitPane(true) {
		t.Fatal("split was rejected")
	}
	stopPaneStartup(a)
	closed := a.active
	closed.working = true
	a.closePane()
	if len(paneLeaves(a.panes)) != 1 || a.active.id != "root" {
		t.Fatal("close did not collapse to the remaining pane")
	}
	if a.sessions[closed.id] != closed || closed.runCtx.Err() != nil {
		t.Fatal("closing a pane stopped or removed its running session")
	}
}

func TestSplitRejectsUnusableTerminalSize(t *testing.T) {
	a := paneTestApp(t, 28, 12)
	before := a.panes
	if a.splitPane(true) || a.panes != before || len(a.sessions) != 1 {
		t.Fatal("unusable split changed the workspace")
	}
}
