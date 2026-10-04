package ui

import (
	"os"
	"strings"

	"jin/internal/core"
	"jin/internal/tools"
)

// showDiff is /diff: the diff of the last turn's edit and write changes in
// the editor. /diff session covers every turn of the session.
func (a *app) showDiff(arg string) {
	s := a.active
	arg = strings.ToLower(strings.TrimSpace(arg))
	if arg != "" && arg != "session" {
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "usage: /diff or /diff session"})
		return
	}
	if !s.persisted {
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Nothing to diff"})
		return
	}
	var changes []tools.Change
	var err error
	if arg == "session" {
		changes, err = s.store.AllChanges(s.id)
	} else {
		_, changes, err = s.store.LastChanges(s.id)
	}
	if err != nil {
		s.persistenceError("Diff failed", err)
		return
	}
	if len(changes) == 0 {
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Nothing to diff. " + diffBashNote})
		return
	}
	file, err := os.CreateTemp("", "jin-diff-*.diff")
	if err != nil {
		a.report(err)
		return
	}
	_, writeErr := file.WriteString(diffText(changes))
	closeErr := file.Close()
	if err := firstErr(writeErr, closeErr); err != nil {
		os.Remove(file.Name())
		a.report(err)
		return
	}
	a.report(a.editFile(file.Name(), func() error { return os.Remove(file.Name()) }))
}
