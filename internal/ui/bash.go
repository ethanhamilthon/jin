package ui

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
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

func (a *app) runBash(s *chatSession, command string) {
	b := s.bash
	b.running = true
	ctx, cancel := context.WithCancel(a.ctx)
	b.cancel = cancel
	s.closeOpenEntry()
	s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "$ " + command})
	s.scroll = 0
	dir, id := a.dir, s.id
	go func() {
		defer cancel()
		ctx, stop := context.WithTimeout(ctx, 10*time.Minute)
		defer stop()
		cmd := exec.CommandContext(ctx, "sh", "-c", command)
		cmd.Dir = dir
		cmd.WaitDelay = 2 * time.Second
		out, err := cmd.CombinedOutput()
		text := strings.TrimRight(string(out), "\n")
		if len(text) > maxBashOutput {
			text = text[:maxBashOutput] + "\n… output cut at 16 KB"
		}
		select {
		case a.bashDone <- bashResult{session: id, command: command, output: text, err: err}:
		case <-a.ctx.Done():
		}
	}()
}

func (a *app) receiveBash(r bashResult) {
	s, ok := a.sessions[r.session]
	if !ok {
		return
	}
	if s.bash != nil {
		s.bash.running, s.bash.cancel = false, nil
	}
	text := r.output
	kind := core.UpdateInfo
	if r.err != nil {
		kind = core.UpdateError
		if text != "" {
			text += "\n"
		}
		text += r.err.Error()
	}
	if text == "" {
		text = "(no output)"
		kind = core.UpdateInfo
	}
	s.closeOpenEntry()
	s.appendEntry(chatEntry{kind: kind, text: text})
	s.scroll = 0
}
