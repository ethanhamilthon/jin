package ui

import (
	"context"
	"jin/internal/core"
	"jin/internal/tools"
	"os/exec"
	"strings"
	"time"
)

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
		tools.Isolate(cmd)
		cmd.Cancel = func() error { tools.KillGroup(cmd); return nil }
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
