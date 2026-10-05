package tasks

import (
	"fmt"
	"slices"
	"sync"
)

// Manager owns the tasks of every agent in this process. Owner is the id of
// the session (or run) whose agent started a task.
type Manager struct {
	mu     sync.Mutex
	tasks  map[string]*task
	events chan Event
}

var shared = New()

// Shared is the manager of this process.
func Shared() *Manager { return shared }

// New makes an empty manager.
func New() *Manager {
	return &Manager{tasks: map[string]*task{}, events: make(chan Event, 256)}
}

// Events delivers one Event per task that ended.
func (m *Manager) Events() <-chan Event { return m.events }

func (m *Manager) get(id string) (*task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return nil, fmt.Errorf("no task %s", id)
	}
	return t, nil
}

func (m *Manager) owned(owner, id string) (*task, error) {
	t, err := m.get(id)
	if err == nil && t.info.Owner != owner {
		return nil, fmt.Errorf("no task %s in this session", id)
	}
	return t, err
}

// List returns the tasks of an owner, oldest first; an empty owner lists all.
func (m *Manager) List(owner string) []Info {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Info
	for _, t := range m.tasks {
		if owner == "" || t.info.Owner == owner {
			out = append(out, t.info)
		}
	}
	slices.SortFunc(out, func(a, b Info) int { return a.Started.Compare(b.Started) })
	return out
}

// Running lists the tasks of an owner that still run.
func (m *Manager) Running(owner string) []Info {
	var out []Info
	for _, info := range m.List(owner) {
		if info.Status == Running {
			out = append(out, info)
		}
	}
	return out
}

// Counts is the number of running tasks per owner.
func (m *Manager) Counts() map[string]int {
	counts := map[string]int{}
	for _, info := range m.Running("") {
		counts[info.Owner]++
	}
	return counts
}
