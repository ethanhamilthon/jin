package ui

import (
	"os"
	"strings"

	"jin/internal/core"
)

// showDiff is /diff: the diff of the last turn's edit and write changes in
// the editor. The whole-session variant needs a store query that does not
// exist yet, so /diff session says so.
func (a *app) showDiff(arg string) {
	s := a.active
	if strings.TrimSpace(arg) != "" {
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "/diff session is not available yet; /diff shows the last turn"})
		return
	}
	if !s.persisted {
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Nothing to diff"})
		return
	}
	turn, changes, err := s.store.LastChanges(s.id)
	if err != nil {
		s.persistenceError("Diff failed", err)
		return
	}
	if turn == 0 {
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
