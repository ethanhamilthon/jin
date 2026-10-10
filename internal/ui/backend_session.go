package ui

import (
	"jin/internal/core"
	"jin/internal/session"
	"jin/internal/tools"
)

func (a *app) backendSession(snap session.Snapshot) *chatSession {
	s := a.sessions[snap.State.ID]
	if s == nil {
		client, _, _ := a.clientFor(snap.State.Provider)
		s = &chatSession{backend: a.backend, id: snap.State.ID, store: a.store, client: client, width: a.width, pricing: a.pricing, fold: a.fold, toolNames: tools.Without(a.cfg.ToolsDisabled)}
		a.sessions[s.id] = s
	}
	s.remoteEntries, s.remoteMap, s.remoteSeq = snap.Entries, nil, snap.Seq
	a.backendState(s, snap.State)
	s.history, s.openKind = nil, ""
	if snap.Intro != nil {
		s.history = a.introEntriesAt(s.path)
	}
	for _, entry := range snap.Entries {
		s.remoteMap = append(s.remoteMap, len(s.history))
		s.history = append(s.history, backendEntry(entry))
	}
	s.rebuildRows(s.width)
	if snap.State.Draft != "" && len(s.input) == 0 {
		s.input = clusters(snap.State.Draft)
		s.cursor = len(s.input)
	}
	return s
}

// remoteRow is where a daemon entry sits in the chat history: entries the
// client adds itself, such as an error row, must not shift the mapping.
func (s *chatSession) remoteRow(index int) (int, bool) {
	if index < 0 || index >= len(s.remoteMap) {
		return 0, false
	}
	return s.remoteMap[index], true
}

func backendEntry(entry session.Entry) chatEntry { return chatEntryOf(entry) }

func (a *app) backendState(s *chatSession, st session.State) {
	old := s.remoteState
	if old.Provider != st.Provider {
		client, _, _ := a.clientFor(st.Provider)
		s.client = client
	}
	s.remoteState = st
	s.path, s.title, s.provider, s.model, s.effort = st.Path, st.Title, st.Provider, st.Model, st.Effort
	s.persisted, s.ready, s.working, s.unread, s.usage = st.Persisted, st.Ready, st.Working, st.Unread, st.Usage
	s.providerMissing = st.ProviderMissing
	if a.tasksRunning == nil {
		a.tasksRunning = map[string]int{}
	}
	a.tasksRunning[s.id] = st.Tasks
	s.inflight = 0
	if st.Busy {
		s.inflight = 1
	}
	if st.Cache != nil {
		s.cache = cacheRate{known: true, percent: *st.Cache}
	}
	if len(st.Ask) == 0 {
		s.ask = nil
	} else if s.ask == nil || old.Question != st.Question {
		s.ask = newAskState(st.Ask)
	}
	if s.bash != nil {
		s.bash.running = st.Shell
	}
}

func (a *app) backendError(err error) {
	if err != nil && a.active != nil {
		a.active.appendEntry(chatEntry{kind: core.UpdateError, text: err.Error()})
	}
}
