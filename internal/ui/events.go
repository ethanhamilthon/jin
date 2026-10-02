package ui

import (
	"github.com/gdamore/tcell/v3"
)

func (a *app) handleEvent(event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventResize:
		a.screen.Sync()
		a.width, _ = a.screen.Size()
	case *tcell.EventFocus:
		a.blurred = !ev.Focused
	case *tcell.EventPaste:
		a.pasting = ev.Start()
	case *tcell.EventMouse:
		a.active.handleMouse(ev, a.screen)
	case *tcell.EventKey:
		if !ev.Pressed() {
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
		case isTodoKey(ev) && a.sel == nil:
			a.editTodos(a.active)
		case isCtrl(ev, 'c', false):
			a.interrupt()
		case a.sel == nil && a.active.ask != nil:
			a.askKey(ev)
		case a.sel != nil:
			a.selectorKey(ev)
		default:
			a.insertKey(ev)
		}
		a.refreshMention()
		a.refreshSlash()
		a.refreshFile()
	}
}

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
	if a.active.bash != nil {
		a.bashKey(ev)
		return
	}
	if !a.slashKey(ev) && !a.fileKey(ev) && !a.mentionKey(ev) {
		a.typeKey(ev)
	}
}

// interrupt is Ctrl+C: stop a running shell command, otherwise the request.
func (a *app) interrupt() {
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
		if text, ok := pasteClipboard(); ok {
			insertClusters(&s.input, &s.cursor, text)
		}
	case ev.Key() == tcell.KeyUp || ev.Key() == tcell.KeyDown:
		delta := 1
		if ev.Key() == tcell.KeyUp {
			delta = -1
		}
		s.cursor = moveVertical(s.input, s.cursor, a.width-2, delta)
	case ev.Key() == tcell.KeyEnter && (a.pasting || ev.Modifiers()&(tcell.ModShift|tcell.ModAlt) != 0):
		insertClusters(&s.input, &s.cursor, "\n")
	case ev.Key() == tcell.KeyEnter && a.runInlineCommand():
	default:
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
