package session

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/store"
)

func readOnlyText(pid int) string { return "Read-only: in use by process " + strconv.Itoa(pid) }

// lockOut keeps a session from reaching its agent: the prompt channel has
// no reader.
func (s *Session) lockOut(pid int) {
	s.readOnlyPID = pid
	s.prompts = make(chan core.Request)
}

func (s *Session) makeReadOnly(pid int) {
	s.lockOut(pid)
	s.add(Entry{Kind: core.UpdateInfo, Text: readOnlyText(pid)})
}

// sendRefusal says why the session cannot take a request. It claims the
// session first, so a process that took it over turns it read-only before
// anything is queued or saved.
func (s *Session) sendRefusal() error {
	projectErr := s.projectError()
	switch {
	case s.providerMissing:
		return errors.New(missingProviderText(s.provider))
	case s.readOnlyPID != 0:
		return errors.New(readOnlyText(s.readOnlyPID))
	case s.model == "":
		return errors.New("Choose a model from an enabled provider first")
	case !s.client.Config().Ready():
		return errors.New("Provider is not ready: pick one in Providers")
	case projectErr != nil:
		return projectErr
	case s.render != nil && s.render.reload:
		return errors.New("The session is reloading prompts; wait for it to finish")
	case !s.persisted:
		return nil
	}
	var busy store.ErrSessionBusy
	if err := s.m.db.SetRunning(s.id, true); err != nil {
		if errors.As(err, &busy) {
			s.lockOut(busy.PID)
			return errors.New(readOnlyText(busy.PID))
		}
		return err
	}
	return nil
}

func (s *Session) projectError() error {
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

func (m *Manager) historyOf(messages []provider.Message) []Entry {
	return History(messages, m.registry)
}

// busyOwner returns the pid of another live jin process that owns the
// session, or 0 when nobody does.
func (m *Manager) busyOwner(id string) int {
	pid, alive, err := m.db.SessionOwner(id)
	if err != nil || !alive || pid == os.Getpid() {
		return 0
	}
	return pid
}
