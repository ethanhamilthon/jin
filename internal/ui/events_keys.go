package ui

import "github.com/gdamore/tcell/v3"

// loadingKey handles a key while the session starts. The input is closed:
// nothing is typed, sent or opened. Ctrl+C stops the commands that still
// run, so a slow one does not hold the session back; Ctrl+O and copying
// still work.
func (a *app) loadingKey(ev *tcell.EventKey) {
	switch {
	case isCopyKey(ev) && a.active.selection.active:
		a.copySelection()
	case isCtrl(ev, 'c', false):
		a.cancelRender(a.active)
	case isFoldKey(ev):
		a.cycleFold()
	}
}

func (a *app) insertKey(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyTab && a.tabSwitchesPane() {
		a.cyclePaneFocus()
		return
	}
	if a.active.bashInput() {
		a.bashKey(ev)
		return
	}
	if !a.slashKey(ev) && !a.fileKey(ev) && !a.mentionKey(ev) {
		a.typeKey(ev)
	}
}

// tabSwitchesPane reports whether Tab belongs to the panes: no autocomplete
// list is open, or the open one has nothing to complete.
func (a *app) tabSwitchesPane() bool {
	p := a.panel()
	return p == nil || len(p.options) == 0
}

// cyclePaneFocus moves the focus to the next pane in layout order, wrapping
// at the end. With one pane it does nothing.
func (a *app) cyclePaneFocus() {
	leaves := paneLeaves(a.panes)
	if len(leaves) < 2 {
		return
	}
	current := a.focusedLeaf()
	next := leaves[0]
	for i, leaf := range leaves {
		if leaf == current {
			next = leaves[(i+1)%len(leaves)]
			break
		}
	}
	if next.session != nil {
		a.focus(next.session)
	}
}

// interrupt is Ctrl+C: stop a running shell command, otherwise the request.
func (a *app) interrupt() {
	if render := a.active.render; render != nil && render.reload {
		render.cancel()
		return
	}
	if b := a.active.bash; b != nil && b.cancel != nil {
		b.cancel()
		return
	}
	a.active.agent.Interrupt()
}

func (a *app) typeKey(ev *tcell.EventKey) {
	s := a.active
	switch {
	case ev.Key() == tcell.KeyEscape:
		// Esc only closes things that are open; with nothing open it does nothing.
	case isPasteKey(ev):
		a.pasteKey()
	case ev.Key() == tcell.KeyUp || ev.Key() == tcell.KeyDown:
		delta := 1
		if ev.Key() == tcell.KeyUp {
			delta = -1
		}
		s.cursor = moveVertical(s.input, s.cursor, a.width-2, delta)
	case ev.Key() == tcell.KeyEnter && (a.pasting || ev.Modifiers()&(tcell.ModShift|tcell.ModAlt) != 0):
		insertClusters(&s.input, &s.cursor, "\n")
	case ev.Key() == tcell.KeyEnter && !s.bashInput() && a.runInlineCommand():
	default:
		a.tokenizeCommand(ev)
		if text := handleInput(ev, &s.input, &s.cursor); text != "" {
			a.sendDraft(text)
		}
	}
}

func (a *app) copySelection() {
	s := a.active
	if text := selectedText(s.rows, s.selection); text != "" {
		copySelection(a.screen, text)
	}
	s.selection = textSelection{}
}

func isCopyKey(ev *tcell.EventKey) bool { return isCtrl(ev, 'c', true) }

func isPasteKey(ev *tcell.EventKey) bool { return isCtrl(ev, 'v', true) }
