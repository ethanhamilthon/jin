package ui

import (
	"strings"
	"time"

	"jin/internal/core"
)

// appendDelta streams a fragment into the currently open entry, starting a
// new one when the kind changes. Rows are rendered later, at most once per
// frame, by flushStream.
func (s *chatSession) appendDelta(kind core.UpdateKind, text string) {
	if s.openKind != "" && s.openKind != kind {
		s.trimOpenEntry()
	}
	if s.openKind == "" {
		s.stream.attempt = streamMark{history: len(s.history), rows: len(s.rows)}
	}
	if s.openKind != kind {
		if s.fold.shows(kind) && needsGap(s.lastShown(), kind) {
			s.rows = append(s.rows, chatRow{})
			if s.scroll > 0 {
				s.scroll++
			}
		}
		s.openKind = kind
		s.openRowStart = len(s.rows)
		s.stream.flushed = time.Time{}
		s.history = append(s.history, chatEntry{kind: kind})
	}
	s.history[len(s.history)-1].text += text
	s.stream.dirty = true
	s.flushStream()
}

// closeOpenEntry ends the current stream, trimming stray trailing
// whitespace/blank lines some models leave at the end of a block.
func (s *chatSession) closeOpenEntry() {
	s.trimOpenEntry()
	s.openKind = ""
}

func (s *chatSession) trimOpenEntry() {
	if s.openKind == "" {
		return
	}
	entry := &s.history[len(s.history)-1]
	trimmed := strings.TrimSpace(entry.text)
	if trimmed == entry.text && !s.stream.dirty {
		return
	}
	entry.text = trimmed
	s.renderOpenEntry()
}

func (s *chatSession) rebuildRows(width int) {
	s.rows = s.rows[:0]
	s.stream.dirty = false
	var prev *chatEntry
	for i := range s.history {
		entry := &s.history[i]
		open := i == len(s.history)-1 && s.openKind != ""
		if !s.fold.shows(entry.kind) {
			if open {
				s.openRowStart = len(s.rows)
			}
			continue
		}
		if needsGap(prev, entry.kind) {
			s.rows = append(s.rows, chatRow{})
		}
		if open {
			s.openRowStart = len(s.rows)
		}
		s.rows = append(s.rows, entryRows(*entry, width)...)
		prev = entry
	}
}
