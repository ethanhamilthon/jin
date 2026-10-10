package ui

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"jin/internal/core"
	"jin/internal/store"
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

// lockOut keeps a session from reaching its agent: the prompt channel has no
// reader.
func (s *chatSession) lockOut(pid int) {
	s.readOnlyPID = pid
	s.prompts = make(chan core.Request)
}

func readOnlyText(pid int) string { return "Read-only: in use by process " + strconv.Itoa(pid) }

func (s *chatSession) makeReadOnly(pid int) {
	s.lockOut(pid)
	s.closeOpenEntry()
	s.appendEntry(chatEntry{kind: core.UpdateInfo, text: readOnlyText(pid)})
}

// sendRefusal says why the session cannot take a request. It claims the
// session first, so a process that took it over since it was opened turns it
// read-only before anything is queued or saved.
func (s *chatSession) sendRefusal() error {
	if s.backend != nil {
		return s.projectError()
	}
	projectErr := s.projectError()
	switch {
	case s.providerMissing:
		return errors.New(missingProviderText(s.provider))
	case s.readOnlyPID != 0:
		return errors.New(readOnlyText(s.readOnlyPID))
	case s.model == "":
		return errors.New("Choose a model from an enabled provider first")
	case s.client != nil && !s.client.Config().Ready():
		return errors.New("Provider is not ready: pick one with /provider")
	case projectErr != nil:
		return projectErr
	case s.render != nil && s.render.reload:
		return errors.New("The session is reloading prompts; wait for it to finish")
	case !s.persisted:
		return nil
	}
	var busy store.ErrSessionBusy
	if err := s.store.SetRunning(s.id, true); err != nil {
		if errors.As(err, &busy) {
			s.lockOut(busy.PID)
			return errors.New(readOnlyText(busy.PID))
		}
		return err
	}
	return nil
}

func (s *chatSession) projectError() error {
	if s.path == "" {
		return nil
	}
	info, err := os.Stat(s.path)
	if err != nil {
		return fmt.Errorf("Project unavailable: %w", err)
	}
	if !info.IsDir() {
		return errors.New("Project unavailable: path is not a directory")
	}
	return nil
}

// refuseSend shows the reason and gives the draft back when the session
// cannot take a message.
func (a *app) refuseSend(text string) bool {
	s := a.active
	err := a.noEnabledProvider(s)
	if err == nil {
		err = s.sendRefusal()
	}
	if err == nil {
		return false
	}
	s.input = clusters(text)
	s.cursor = len(s.input)
	a.report(err)
	return true
}
