package ui

import (
	"jin/internal/files"
	"os"
	"strings"

	"github.com/gdamore/tcell/v3"
)

// fileKey lets the open list take navigation, completion and Esc. Enter
// completes unless the typed path is already a finished file, so a complete
// @file still sends.
func (a *app) fileKey(ev *tcell.EventKey) bool {
	f := a.file
	if f == nil {
		return false
	}
	switch ev.Key() {
	case tcell.KeyUp:
		f.sel.move(-1)
	case tcell.KeyDown:
		f.sel.move(1)
	case tcell.KeyTab:
		a.acceptFile()
	case tcell.KeyEnter:
		current := f.sel.current()
		complete := !f.dirs[current] && (current == f.typed || current == `"`+f.typed+`"`)
		if a.pasting || ev.Modifiers()&(tcell.ModShift|tcell.ModAlt) != 0 || complete {
			return false
		}
		a.acceptFile()
	case tcell.KeyEscape:
		s := a.active
		a.closed = closedToken{session: s.id, kind: '@', start: f.start, query: f.typed}
		a.file = nil
	default:
		return false
	}
	return true
}

// acceptFile puts the highlighted path after the "@". A directory keeps the
// cursor inside the token so the list goes on into it; a file ends the token.
func (a *app) acceptFile() {
	s, f := a.active, a.file
	value := f.sel.current()
	if value == "" {
		return
	}
	completion := clusters(value)
	isDir := f.dirs[value]
	next := s.cursor
	if f.end > next {
		next = f.end
	}
	tail := s.input[next:]
	cursorAt := f.start + 1 + len(completion)
	if isDir && strings.HasSuffix(value, `"`) {
		cursorAt-- // stay before the closing quote
	}
	if !isDir && (len(tail) == 0 || strings.TrimSpace(tail[0]) != "") {
		completion = append(completion, " ")
		cursorAt = f.start + 1 + len(completion)
	}
	rest := append(completion, tail...)
	s.input = append(s.input[:f.start+1], rest...)
	s.cursor = cursorAt
}

// sendDraft sends the draft. Every typed @path that names a real file is
// replaced by its full path and listed in an <attached-files> block. The chat
// shows token labels; the model gets their content, pasted text literally.
func (a *app) sendDraft(text string) {
	if !a.active.ready {
		return
	}
	if a.refuseSend(text) {
		return
	}
	home, _ := os.UserHomeDir()
	clean, paths := files.Extract(text, home, a.dir)
	shown := renderTokens(text, tokenLabel)
	a.active.sendFiles(shown, renderTokens(clean, tokenModelText), files.Block(paths), promptNames(text))
	a.pruneTokens()
}
