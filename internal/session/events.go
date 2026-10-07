package session

import "jin/internal/core"

// Event is what front ends hear about. Type says which fields are set:
//
//	entry    Index, Entry: a new entry
//	update   Index, Entry: an entry that changed
//	delta    Index, Text: text streamed into an entry
//	entries  Entries: all entries again
//	state    State
//	ring     Kind: "done" or "ask", for the notification sound
//	handoff  Text: the new session the brief went to
//	sessions, tasks, config, projects: a list changed; fetch it again
type Event struct {
	Type    string          `json:"type"`
	Session string          `json:"session,omitempty"`
	Index   int             `json:"index,omitempty"`
	Entry   *Entry          `json:"entry,omitempty"`
	Entries []Entry         `json:"entries,omitempty"`
	Text    string          `json:"text,omitempty"`
	Kind    core.UpdateKind `json:"kind,omitempty"`
	State   *State          `json:"state,omitempty"`
	Seq     int64           `json:"seq"`
}

func (s *Session) emit(ev Event) {
	ev.Session = s.id
	s.m.publish(ev)
}

// publish numbers an event; a snapshot carries the number of the last
// event before it, so a front end skips what the snapshot already holds.
func (m *Manager) publish(ev Event) {
	ev.Seq = m.seq.Add(1)
	m.out(ev)
}

func (s *Session) emitState() {
	state := s.state()
	s.emit(Event{Type: "state", State: &state})
}

func (s *Session) emitEntries() {
	s.emit(Event{Type: "entries", Entries: append([]Entry{}, s.entries...)})
}
