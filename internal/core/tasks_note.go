package core

import (
	"fmt"
	"strings"
	"time"
)

const tasksNoteTag = "<background-tasks>"

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
	lines := []string{tasksNoteTag, "Still running:"}
	for _, info := range running {
		lines = append(lines, fmt.Sprintf("- %s %q (%s)", info.ID, info.Command, time.Since(info.Started).Round(time.Second)))
	}
	if a.tasksEndWithRun {
		lines = append(lines, "They are stopped when you finish. If you need a result, check the task before you finish; otherwise reply in one short line.")
	} else {
		lines = append(lines, "They keep running after you finish, and their results arrive later as messages. Stop the ones that are no longer needed with the task tool, then reply in one short line. If they all should keep running, reply in one short line saying so.")
	}
	return strings.Join(append(lines, "</background-tasks>"), "\n"), true
}

// IsTasksNote reports a message made by tasksNote.
func IsTasksNote(text string) bool { return strings.HasPrefix(text, tasksNoteTag) }
