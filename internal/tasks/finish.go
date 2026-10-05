package tasks

import (
	"errors"
	"os/exec"
)

// resultChars is how much of the end of the output an Event carries.
const resultChars = 8000

func (m *Manager) finish(t *task, err error) {
	code := exitCode(err)
	m.mu.Lock()
	t.closeStdin()
	t.info.Exit = code
	switch {
	case t.stopBy != "":
		t.info.Status = Stopped
	case code == 0:
		t.info.Status = Done
	default:
		t.info.Status = Failed
	}
	info, by := t.info, t.stopBy
	m.mu.Unlock()
	close(t.ended)
	body, _, _ := tail(info.Log, resultChars)
	if by != "" {
		body = "stopped " + by + "\n" + body
	}
	if by == ByAgent || by == byExit {
		return
	}
	event := Event{Owner: info.Owner, Task: info, Text: ResultText(info.ID, info.Status, info.Exit, body)}
	go func() { m.events <- event }()
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		if code := exit.ExitCode(); code >= 0 {
			return code
		}
		return signalCode(exit)
	}
	return 1
}
