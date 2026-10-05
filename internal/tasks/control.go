package tasks

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"jin/internal/tasklog"
)

// Who stopped a task; the agent and jin's own exit get no Event.
const (
	ByAgent = "by the agent"
	ByUser  = "manually by the user"
	byExit  = "because jin exited"
)

// inputTimeout bounds one Input: waiting for an earlier input of the same
// task and writing to a child that does not read.
var inputTimeout = 5 * time.Second

// Check returns the task and the last limit characters of its output;
// limit <= 0 returns all of it.
func (m *Manager) Check(owner, id string, limit int) (Info, string, error) {
	t, err := m.owned(owner, id)
	if err != nil {
		return Info{}, "", err
	}
	m.mu.Lock()
	info := t.info
	m.mu.Unlock()
	text, _, err := tail(info.Log, limit)
	return info, text, err
}

// Input writes text to the stdin of a task started with stdin.
func (m *Manager) Input(owner, id, text string) error {
	t, err := m.owned(owner, id)
	if err != nil {
		return err
	}
	if t.stdin == nil {
		return fmt.Errorf("task %s has no stdin; start the task with stdin to write to it", id)
	}
	select {
	case <-t.ended:
		return fmt.Errorf("task %s is not running", id)
	case t.writing <- struct{}{}:
		defer func() { <-t.writing }()
	case <-time.After(inputTimeout):
		return fmt.Errorf("task %s is still busy with an earlier input", id)
	}
	_ = t.stdin.SetWriteDeadline(time.Now().Add(inputTimeout))
	_, err = io.WriteString(t.stdin, text)
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return fmt.Errorf("task %s does not read its stdin; the input was not fully written", id)
	}
	return err
}

// Stop ends a running task of owner: SIGTERM to its group, SIGKILL after 3 s.
func (m *Manager) Stop(owner, id, by string) error {
	t, err := m.owned(owner, id)
	if err != nil {
		return err
	}
	return m.stop(t, by)
}

// StopAny is Stop for the user, who may stop the task of any session.
func (m *Manager) StopAny(id string) error {
	t, err := m.get(id)
	if err != nil {
		return err
	}
	return m.stop(t, ByUser)
}

func (m *Manager) stop(t *task, by string) error {
	m.mu.Lock()
	if t.info.Status != Running {
		m.mu.Unlock()
		return fmt.Errorf("task %s is not running", t.info.ID)
	}
	t.stopBy = by
	m.mu.Unlock()
	if by == byExit {
		killNow(t.cmd)
	} else {
		terminate(t.cmd, t.ended)
	}
	return nil
}

// StopAll kills every running task. Call it when jin exits.
func (m *Manager) StopAll() {
	for _, info := range m.Running("") {
		if t, err := m.get(info.ID); err == nil {
			_ = m.stop(t, byExit)
		}
	}
}

func tail(path string, limit int) (string, bool, error) { return tasklog.Tail(path, limit) }
