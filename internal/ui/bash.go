package ui

import (
	"context"
	"strings"

	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
)

// maxBashOutput caps what one shell command may show in the timeline.
const maxBashOutput = 16 << 10

// bashState tracks execution separately from the session's input draft.
type bashState struct {
	running bool
	cancel  context.CancelFunc
	done    <-chan struct{}
}

type bashResult struct {
	session string
	command string
	output  string
	err     error
}

// bashInput checks the rendered draft so paste tokens and paste markers count.
func (s *chatSession) bashInput() bool {
	if s == nil {
		return false
	}
	text := renderTokens(strings.Join(s.input, ""), func(t inputToken) string {
		if t.kind == tokenPaste {
			return t.payload
		}
		return t.label
	})
	return strings.HasPrefix(text, "$")
}

func bashCommand(input []string) (string, bool) {
	text := draftPayload(input)
	if !strings.HasPrefix(text, "$") {
		return "", false
	}
	command := strings.TrimPrefix(text, "$")
	if strings.TrimSpace(command) == "" {
		return "", false
	}
	return command, true
}

// bashKey edits and submits the ordinary draft while its first character is $.
func (a *app) bashKey(ev *tcell.EventKey) {
	s := a.active
	switch {
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
	case ev.Key() == tcell.KeyEnter:
		a.submitBash(s)
	case ev.Key() == tcell.KeyEscape:
	case ev.Key() == tcell.KeyRune || ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 || ev.Key() == tcell.KeyDelete || ev.Key() == tcell.KeyLeft || ev.Key() == tcell.KeyRight || ev.Key() == tcell.KeyHome || ev.Key() == tcell.KeyEnd:
		handleInput(ev, &s.input, &s.cursor)
	}
}

func (a *app) submitBash(s *chatSession) {
	command, ok := bashCommand(s.input)
	if !ok {
		return
	}
	if s.bash != nil && s.bash.running {
		s.appendEntry(chatEntry{kind: core.UpdateError, text: "A shell command is already running."})
		return
	}
	if s.readOnlyPID != 0 {
		s.appendEntry(chatEntry{kind: core.UpdateError, text: readOnlyText(s.readOnlyPID)})
		return
	}
	if !a.runBash(s, command) {
		return
	}
	s.input, s.cursor, s.inputTop = nil, 0, 0
	a.pruneTokens()
}
