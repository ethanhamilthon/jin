package session

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"

	"jin/internal/core"
	"jin/internal/store"
	"jin/internal/tools"
)

// maxShellOutput caps what one shell command may show.
const maxShellOutput = 16 << 10

// Shell runs a command the user typed after $ in the session's directory;
// its output shows in the chat and the agent does not see it.
func (m *Manager) Shell(id, command string) error {
	return m.Do(id, func(s *Session) error {
		switch {
		case strings.TrimSpace(command) == "":
			return nil
		case s.shell != nil:
			return errors.New("A shell command is already running.")
		case s.readOnlyPID != 0:
			return errors.New(readOnlyText(s.readOnlyPID))
		}
		if err := s.projectError(); err != nil {
			return err
		}
		if s.persisted {
			var busy store.ErrSessionBusy
			if err := m.db.SetRunning(s.id, true); errors.As(err, &busy) {
				s.makeReadOnly(busy.PID)
				return errors.New(readOnlyText(busy.PID))
			}
		}
		ctx, cancel := context.WithTimeout(s.runCtx, 10*time.Minute)
		s.shell = cancel
		s.add(Entry{Kind: core.UpdateInfo, Tool: "shell", Text: "$ " + command})
		s.emitState()
		go m.runShell(ctx, cancel, s, command)
		return nil
	})
}

func (m *Manager) runShell(ctx context.Context, cancel context.CancelFunc, s *Session, command string) {
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = s.path
	cmd.WaitDelay = 2 * time.Second
	tools.Isolate(cmd)
	cmd.Cancel = func() error { tools.KillGroup(cmd); return nil }
	out, err := cmd.CombinedOutput()
	text := strings.TrimRight(string(out), "\n")
	if len(text) > maxShellOutput {
		text = text[:maxShellOutput] + "\n… output cut at 16 KB"
	}
	kind := core.UpdateInfo
	if err != nil {
		kind = core.UpdateError
		text = strings.TrimLeft(text+"\n"+err.Error(), "\n")
	}
	if text == "" {
		text, kind = "(no output)", core.UpdateInfo
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s.shell = nil
	if s.persisted && !s.working && s.inflight == 0 && len(s.pending) == 0 {
		_ = m.db.SetRunning(s.id, false)
	}
	s.add(Entry{Kind: kind, Tool: "shell", Text: text})
	s.emitState()
}
