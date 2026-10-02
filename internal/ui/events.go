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
		switch {
		case isCopyKey(ev) && a.active.selection.active:
			a.copySelection()
		case isCycleModelKey(ev) && a.sel == nil:
			a.cycleModel()
		case isFoldKey(ev) && a.sel == nil:
			a.cycleFold()
		case isTodoKey(ev) && a.sel == nil:
			a.editTodos(a.active)
		case ev.Key() == tcell.KeyCtrlC:
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

func isCopyKey(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyCtrlC {
		return true
	}
	return ev.Key() == tcell.KeyRune && ev.Str() == "c" && ev.Modifiers()&(tcell.ModCtrl|tcell.ModMeta) != 0
}

func isPasteKey(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyCtrlV {
		return true
	}
	return ev.Key() == tcell.KeyRune && ev.Str() == "v" && ev.Modifiers()&(tcell.ModCtrl|tcell.ModMeta) != 0
}
