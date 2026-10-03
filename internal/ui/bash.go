package ui

import (
	"context"
	"strings"

	"github.com/gdamore/tcell/v3"
)

// maxBashOutput caps what one /bash command may show in the chat.
const maxBashOutput = 16 << 10

// bashState is the /bash mode of one session. It has its own input line, so
// the chat draft stays as it was. Output goes to the chat only; it is never
// sent to the model.
type bashState struct {
	input   []string
	cursor  int
	top     int
	running bool
	cancel  context.CancelFunc
}

type bashResult struct {
	session string
	command string
	output  string
	err     error
}

func (a *app) startBash() {
	a.slash, a.mention, a.file = nil, nil, nil
	if a.active.bash == nil {
		a.active.bash = &bashState{}
	}
}

// bashKey handles a key while the shell input is open.
func (a *app) bashKey(ev *tcell.EventKey) {
	s := a.active
	b := s.bash
	switch {
	case ev.Key() == tcell.KeyEscape:
		if b.cancel != nil {
			b.cancel()
		}
		s.bash = nil
	case isPasteKey(ev):
		if text, ok := pasteClipboard(); ok {
			insertClusters(&b.input, &b.cursor, strings.ReplaceAll(text, "\n", " "))
		}
	case ev.Key() == tcell.KeyUp || ev.Key() == tcell.KeyDown:
	case ev.Key() == tcell.KeyEnter && b.running:
	default:
		if command := handleInput(ev, &b.input, &b.cursor); command != "" {
			a.runBash(s, command)
		}
	}
}
