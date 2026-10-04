package ui

import "time"

// streamFrame is the shortest time between two renders of a streaming entry.
const streamFrame = 50 * time.Millisecond

// streamMark is where the entries of the current model attempt begin.
type streamMark struct{ history, rows int }

type streamState struct {
	attempt streamMark
	dirty   bool
	flushed time.Time
	renders int
}

// flushStream renders the open entry when it changed and a frame has passed.
// The draw loop calls it, so text that arrives last shows within one tick.
func (s *chatSession) flushStream() {
	if s.stream.dirty && time.Since(s.stream.flushed) >= streamFrame {
		s.renderOpenEntry()
	}
}

func (s *chatSession) renderOpenEntry() {
	rows := s.entryRows(s.history[len(s.history)-1])
	if s.scroll > 0 {
		s.scroll = max(0, s.scroll+len(rows)-(len(s.rows)-s.openRowStart))
	}
	s.rows = append(s.rows[:s.openRowStart], rows...)
	s.stream.dirty, s.stream.flushed = false, time.Now()
	s.stream.renders++
}

// dropAttempt removes what a failed attempt streamed. The stored history
// never had it, so only the screen changes.
func (s *chatSession) dropAttempt() {
	if s.openKind == "" {
		return
	}
	mark := s.stream.attempt
	if s.scroll > 0 {
		s.scroll = max(0, s.scroll-(len(s.rows)-mark.rows))
	}
	s.history = s.history[:mark.history]
	s.rows = s.rows[:mark.rows]
	s.openKind = ""
	s.stream.dirty = false
}
