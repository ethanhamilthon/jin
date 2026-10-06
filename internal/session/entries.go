package session

import (
	"slices"
	"strings"

	"jin/internal/core"
)

// add closes the streaming entry and appends one.
func (s *Session) add(entry Entry) {
	s.closeOpen()
	s.joinNext = false
	s.entries = append(s.entries, entry)
	s.emit(Event{Type: "entry", Index: len(s.entries) - 1, Entry: &entry})
}

// appendDelta streams a fragment into the open entry, starting a new one
// when the kind changes.
func (s *Session) appendDelta(kind core.UpdateKind, text string) {
	if s.joinNext {
		if kind == core.UpdateReasoning {
			return
		}
		s.joinNext = false
		if last := len(s.entries) - 1; last >= 0 && s.entries[last].Kind == core.UpdateAssistant {
			s.entries[last].Text += "\n\n"
			s.open = core.UpdateAssistant
			s.emitUpdate(last)
		}
	}
	if s.open != kind {
		s.closeOpen()
		s.open = kind
		s.entries = append(s.entries, Entry{Kind: kind})
		s.attempt = append(s.attempt, len(s.entries)-1)
		last := len(s.entries) - 1
		s.emit(Event{Type: "entry", Index: last, Entry: &Entry{Kind: kind}})
	}
	last := len(s.entries) - 1
	s.entries[last].Text += text
	s.emit(Event{Type: "delta", Index: last, Text: text})
}

// closeOpen ends the stream, trimming whitespace some models leave at the
// end of a block.
func (s *Session) closeOpen() {
	if s.open == "" {
		return
	}
	s.open = ""
	last := len(s.entries) - 1
	trimmed := strings.TrimSpace(s.entries[last].Text)
	if trimmed != s.entries[last].Text {
		s.entries[last].Text = trimmed
		s.emitUpdate(last)
	}
}

func (s *Session) emitUpdate(i int) {
	entry := s.entries[i]
	s.emit(Event{Type: "update", Index: i, Entry: &entry})
}

// dropAttempt removes what a failed attempt streamed; the stored history
// never had it.
func (s *Session) dropAttempt() {
	failed := s.attempt
	s.attempt, s.open = nil, ""
	if len(failed) == 0 {
		return
	}
	var kept []Entry
	for i, entry := range s.entries {
		if !slices.Contains(failed, i) {
			kept = append(kept, entry)
		}
	}
	s.entries = kept
	s.emitEntries()
}
