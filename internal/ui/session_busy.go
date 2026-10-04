package ui

import (
	"os"
	"strconv"

	"jin/internal/core"
)

// busyOwner returns the pid of another live jin process that owns the
// session, or 0 when nobody does.
func (a *app) busyOwner(id string) int {
	pid, alive, err := a.store.SessionOwner(id)
	if err != nil || !alive || pid == os.Getpid() {
		return 0
	}
	return pid
}

// makeReadOnly keeps the history of a session that another process uses.
// Nothing reaches its agent: the prompt channel has no reader, so a message
// typed here is never sent and never saved.
func (s *chatSession) makeReadOnly(pid int) {
	s.readOnlyPID = pid
	s.prompts = make(chan core.Request)
	s.closeOpenEntry()
	s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Read-only: in use by process " + strconv.Itoa(pid)})
}
