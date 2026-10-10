package session

import (
	"slices"

	"jin/internal/store"
	"jin/internal/tools"
)

// State is everything about a session a front end shows besides entries.
type State struct {
	ID              string           `json:"id"`
	Path            string           `json:"path"`
	Title           string           `json:"title"`
	Model           string           `json:"model"`
	Effort          string           `json:"effort"`
	Provider        string           `json:"provider"`
	Persisted       bool             `json:"persisted"`
	Ready           bool             `json:"ready"`
	Working         bool             `json:"working"`
	Busy            bool             `json:"busy"`
	Unread          bool             `json:"unread"`
	ReadOnly        int              `json:"read_only,omitempty"`
	ProviderMissing bool             `json:"provider_missing,omitempty"`
	Usage           store.Usage      `json:"usage"`
	Cache           *int             `json:"cache,omitempty"`
	Window          int              `json:"window"`
	Ask             []tools.Question `json:"ask,omitempty"`
	Loading         []string         `json:"loading,omitempty"`
	Reloading       bool             `json:"reloading,omitempty"`
	Shell           bool             `json:"shell,omitempty"`
	Queued          int              `json:"queued,omitempty"`
	Paused          bool             `json:"paused,omitempty"`
	Question        int64            `json:"question,omitempty"`
	Tasks           int              `json:"tasks,omitempty"`
	Draft           string           `json:"draft,omitempty"`
	DraftRev        int              `json:"draft_rev,omitempty"`
}

// Snapshot is a session as a front end opens it.
type Snapshot struct {
	State   State   `json:"state"`
	Entries []Entry `json:"entries"`
	Intro   *Intro  `json:"intro,omitempty"`
	Seq     int64   `json:"seq"`
}

func (s *Session) state() State {
	st := State{
		ID: s.id, Path: s.path, Title: s.title, Model: s.model, Effort: s.effort, Provider: s.provider,
		Persisted: s.persisted, Ready: s.ready, Working: s.working, Busy: s.busy(), Unread: s.unread,
		ReadOnly: s.readOnlyPID, ProviderMissing: s.providerMissing, Usage: s.usage, Cache: s.cache,
		Window: s.window(), Ask: s.ask,
		Shell: s.shell != nil, Queued: len(s.pending), Tasks: s.m.tasksRunning(s.id),
		Draft: s.draft, DraftRev: s.draftRev, Paused: s.paused, Question: s.questionRevision,
	}
	if s.render != nil {
		st.Loading, st.Reloading = slices.Clone(s.render.loading), s.render.reload
	}
	return st
}

func (s *Session) snapshot() Snapshot {
	snap := Snapshot{State: s.state(), Entries: append([]Entry{}, s.entries...), Seq: s.m.seq.Load()}
	if !s.persisted {
		snap.Intro = s.intro
	}
	return snap
}

// Snapshot returns a live session as a front end opens it.
func (m *Manager) Snapshot(id string) (Snapshot, error) {
	var snap Snapshot
	err := m.Do(id, func(s *Session) error {
		snap = s.snapshot()
		return nil
	})
	return snap, err
}

// Focus remembers the session as the last one of its project.
func (m *Manager) Focus(id string) error {
	return m.Do(id, func(s *Session) error {
		if s.persisted {
			return m.db.RememberProject(s.path, s.id)
		}
		return nil
	})
}
