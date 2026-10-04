package ui

import (
	"slices"
	"time"
)

// streamFrame is the shortest time between two renders of a streaming entry.
const streamFrame = 50 * time.Millisecond

type streamState struct {
	// attempt lists the history entries the running model attempt streamed,
	// whether or not they are still open.
	attempt []int
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
	failed := s.stream.attempt
	s.endAttempt()
	if len(failed) == 0 {
		return
	}
	var kept []chatEntry
	for i, entry := range s.history {
		if !slices.Contains(failed, i) {
			kept = append(kept, entry)
		}
	}
	rows := len(s.rows)
	s.history, s.openKind, s.stream.dirty = kept, "", false
	s.rebuildRows(s.width)
	if s.scroll > 0 {
		s.scroll = max(0, s.scroll-(rows-len(s.rows)))
	}
}

func (s *chatSession) endAttempt() { s.stream.attempt = nil }
