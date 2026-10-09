package ui

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

func (a *app) handleEvent(event tcell.Event) {
	owner := a.active
	before, revision := strings.Join(owner.input, ""), owner.draftRevision
	defer func() {
		if strings.Join(owner.input, "") != before && owner.draftRevision == revision {
			owner.draftRevision++
		}
	}()
	switch ev := event.(type) {
	case *tcell.EventResize:
		a.screen.Sync()
		a.width, _ = a.screen.Size()
	case *tcell.EventFocus:
		a.blurred = !ev.Focused
	case *tcell.EventPaste:
		a.pasting = ev.Start()
		if a.pasting {
			a.beginPaste()
		} else {
			a.endPaste()
		}
		a.refreshPanels()
	case *tcell.EventMouse:
		if a.sel == nil {
			if !a.paneMouse(ev) && a.panes == nil {
				a.active.handleMouse(ev, a.screen)
			}
		}
	case *tcell.EventKey:
		if !ev.Pressed() {
			return
		}
		if a.sel == nil && !a.pasting && a.paneFocusKey(ev) {
			a.refreshPanels()
			return
		}
		if a.sel == nil && a.onboarding() {
			a.onboardingKey(ev)
			return
		}
		if a.sel == nil && !a.active.ready {
			a.loadingKey(ev)
			return
		}
		switch {
		case isCopyKey(ev) && a.active.selection.active:
			a.copySelection()
		case isCycleModelKey(ev) && a.sel == nil:
			a.cycleModel()
		case isFoldKey(ev) && a.sel == nil:
			a.cycleFold()
		case isCtrl(ev, 'c', false):
			a.interrupt()
		case a.sel == nil && a.active.ask != nil:
			a.askKey(ev)
		case a.sel != nil:
			a.selectorKey(ev)
		default:
			a.insertKey(ev)
		}
		a.refreshPanels()
	}
}

// refreshPanels opens or closes the autocomplete lists for the draft as it is.
func (a *app) refreshPanels() {
	a.refreshMention()
	a.refreshSlash()
	a.refreshFile()
}
