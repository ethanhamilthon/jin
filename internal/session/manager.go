package session

import (
	"context"
	"sync"
	"sync/atomic"

	"jin/internal/pricing"
	"jin/internal/store"
	"jin/internal/tasks"
	"jin/internal/tools"
)

// Manager holds the live sessions of one process. Its methods are safe for
// concurrent use; events go to publish, which must not block.
type Manager struct {
	mu       sync.Mutex
	ctx      context.Context
	db       *store.DB
	version  string
	out      func(Event)
	seq      atomic.Int64
	events   eventBus
	sessions map[string]*Session
	prices   pricing.Table
	held     []tasks.Event
	running  map[string]int
	latest   string
	registry *tools.Registry
}

func NewManager(ctx context.Context, db *store.DB, version string, publish func(Event)) *Manager {
	return &Manager{ctx: ctx, db: db, version: version, out: publish, sessions: map[string]*Session{}, running: map[string]int{},
		registry: tools.Build(tools.Catalog())}
}

// Start takes the prices when they arrive and the results of background
// tasks until ctx ends.
func (m *Manager) Start(prices <-chan pricing.Table, events <-chan tasks.Event) {
	go func() {
		for {
			select {
			case <-m.ctx.Done():
				return
			case table := <-prices:
				m.mu.Lock()
				m.prices = table
				m.mu.Unlock()
				prices = nil
			case ev := <-events:
				m.mu.Lock()
				m.receiveTask(ev)
				m.mu.Unlock()
				m.publish(Event{Type: "tasks"})
			}
		}
	}()
}

func (m *Manager) DB() *store.DB   { return m.db }
func (m *Manager) Version() string { return m.version }
func (m *Manager) Prices() pricing.Table {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.prices
}

// Do runs fn on a live session under the lock.
func (m *Manager) Do(id string, fn func(s *Session) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return errNotOpen(id)
	}
	return fn(s)
}

type errNotOpen string

func (e errNotOpen) Error() string { return "session " + string(e) + " is not open" }

// Live lists the states of the open sessions.
func (m *Manager) Live() []State {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []State
	for _, s := range m.sessions {
		out = append(out, s.state())
	}
	return out
}

func (m *Manager) tasksRunning(id string) int {
	n := 0
	for _, info := range tasks.Shared().Running(id) {
		if info.Owner == id {
			n++
		}
	}
	return n
}
