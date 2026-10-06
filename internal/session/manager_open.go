package session

import (
	"errors"
	"os"

	"jin/internal/core"
	"jin/internal/provider"
)

// Create starts a new session in dir; it is saved with its first message.
func (m *Manager) Create(dir string) (Snapshot, error) {
	cfg, err := m.db.LoadConfig()
	if err != nil {
		return Snapshot{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.start(cfg, dir, NewID(), cfg.ActiveProvider, cfg.Model, cfg.Effort, nil, nil)
	s.intro = m.intro(cfg, dir, s.loadingNames())
	return s.snapshot(), nil
}

// Open starts the backend of a saved session, or returns it when it is
// already open.
func (m *Manager) Open(id string) (Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, err := m.openLocked(id)
	if err != nil {
		return Snapshot{}, err
	}
	return s.snapshot(), nil
}

func (m *Manager) openLocked(id string) (*Session, error) {
	if s, ok := m.sessions[id]; ok {
		return s, nil
	}
	rec, found, err := m.db.GetSession(id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.New("no session with id " + id)
	}
	messages, err := m.db.LoadMessages(rec.ID)
	if err != nil {
		return nil, err
	}
	owner := m.busyOwner(rec.ID)
	if owner == 0 {
		for _, msg := range core.InterruptedToolMessages(messages) {
			if err := m.db.AppendMessage(rec.ID, msg); err != nil {
				return nil, err
			}
			messages = append(messages, msg)
		}
	}
	cfg, err := m.db.LoadConfig()
	if err != nil {
		return nil, err
	}
	s := m.start(cfg, rec.Path, rec.ID, rec.Provider, rec.Model, rec.Effort, core.SinceLastSummary(messages), m.historyOf(messages))
	s.persisted, s.title, s.usage = true, rec.Title, rec.Usage
	if owner != 0 {
		s.makeReadOnly(owner)
	}
	s.agent.SetContextSize(rec.Usage.Context)
	if items, err := m.db.LoadTodos(rec.ID); err == nil {
		s.todos = items
	}
	unread, _ := m.db.AllUnread()
	s.unread = unread[rec.ID]
	m.retryTasks(rec.ID)
	return s, nil
}

// Close stops a session that is not working and forgets it. A working
// session keeps running; it reports false.
func (m *Manager) Close(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return true
	}
	if s.busy() {
		return false
	}
	s.stop()
	delete(m.sessions, id)
	return true
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
