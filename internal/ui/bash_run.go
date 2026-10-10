package ui

import (
	"context"
	"jin/internal/core"
	"jin/internal/daemon"
	"jin/internal/tools"
	"os/exec"
	"strings"
	"time"
)

func (a *app) runBash(s *chatSession, command string) bool {
	if a.backend != nil {
		err := a.backend.Command(a.ctx, daemon.Command{Action: "shell", Session: s.id, Text: command}, nil)
		a.backendError(err)
		return err == nil
	}
	if !a.claimShell(s) {
		return false
	}
	if s.bash == nil {
		s.bash = &bashState{}
	}
	b := s.bash
	b.running = true
	done := make(chan struct{})
	b.done = done
	ctx, cancel := context.WithCancel(a.ctx)
	b.cancel = cancel
	s.closeOpenEntry()
	s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "$ " + command})
	s.scroll = 0
	dir, id := s.path, s.id
	go func() {
		defer close(done)
		defer cancel()
		ctx, stop := context.WithTimeout(ctx, 10*time.Minute)
		defer stop()
		cmd := shellCommand(ctx, dir, command)
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
	return true
}

// shellCommand is the user's ! shell: isolated like the bash tool, and
// cancelling it ends the whole process group.
func shellCommand(ctx context.Context, dir, command string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	cmd.WaitDelay = 2 * time.Second
	tools.Isolate(cmd)
	cmd.Cancel = func() error { tools.KillGroup(cmd); return nil }
	return cmd
}

func (a *app) receiveBash(r bashResult) {
	s, ok := a.sessions[r.session]
	if !ok {
		return
	}
	if s.bash != nil {
		s.bash.running, s.bash.cancel = false, nil
	}
	if s.persisted && s.store != nil && !s.working && s.inflight == 0 && len(s.pending) == 0 {
		_ = s.store.SetRunning(s.id, false)
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
