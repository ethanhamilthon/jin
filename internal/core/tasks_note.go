package core

import (
	"fmt"
	"jin/internal/wire"
	"strings"
	"time"
)

const tasksNoteTag = wire.TasksOpen

// tasksNote is what the agent reads when it answers while its background
// tasks still run: it may stop the ones it no longer needs before the turn ends.
func (a *Agent) tasksNote() (string, bool) {
	if a.tasks == nil {
		return "", false
	}
	running := a.tasks.Running(a.tasksOwner)
	if len(running) == 0 {
		return "", false
	}
	lines := []string{tasksNoteTag, wire.TasksStillRunning}
	for _, info := range running {
		lines = append(lines, fmt.Sprintf("- %s %q (%s)", info.ID, info.Command, time.Since(info.Started).Round(time.Second)))
	}
	if a.tasksEndWithRun {
		lines = append(lines, wire.TasksEndWithRun)
	} else {
		lines = append(lines, wire.TasksKeepRunning)
	}
	return strings.Join(append(lines, wire.TasksClose), "\n"), true
}

// IsTasksNote reports a message made by tasksNote.
func IsTasksNote(text string) bool { return strings.HasPrefix(text, tasksNoteTag) }
