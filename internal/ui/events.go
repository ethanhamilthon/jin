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
		case ev.Key() == tcell.KeyCtrlC:
			a.active.agent.Interrupt()
		case a.sel != nil:
			a.selectorKey(ev)
		case a.mode == modeInsert:
			a.insertKey(ev)
		default:
			a.normalKey(ev)
		}
		a.refreshMention()
	}
}

func (a *app) insertKey(ev *tcell.EventKey) {
	if !a.mentionKey(ev) {
		a.typeKey(ev)
	}
}

func (a *app) typeKey(ev *tcell.EventKey) {
	s := a.active
	switch {
	case ev.Key() == tcell.KeyEscape:
		a.mode = modeNormal
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
	default:
		if text := handleInput(ev, &s.input, &s.cursor); text != "" {
			s.send(text)
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
