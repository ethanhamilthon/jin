package ui

import (
	"strings"

	"jin/internal/core"
)

// appendDelta streams a fragment into the currently open entry, starting a
// new one when the kind changes. It returns how many rows the entry grew by.
func (s *chatSession) appendDelta(kind core.UpdateKind, text string) int {
	if s.openKind != "" && s.openKind != kind {
		s.trimOpenEntry()
	}
	before := 0
	if s.openKind == kind {
		before = len(s.rows) - s.openRowStart
	} else {
		s.openKind = kind
		s.openRowStart = len(s.rows)
		s.history = append(s.history, chatEntry{kind: kind})
	}
	entry := &s.history[len(s.history)-1]
	entry.text += text
	newRows := entryRows(*entry, s.width)
	s.rows = append(s.rows[:s.openRowStart], newRows...)
	return len(newRows) - before
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
	if trimmed == entry.text {
		return
	}
	entry.text = trimmed
	s.rows = append(s.rows[:s.openRowStart], entryRows(*entry, s.width)...)
}

func (s *chatSession) rebuildRows(width int) {
	s.rows = s.rows[:0]
	for i, entry := range s.history {
		if i == len(s.history)-1 && s.openKind != "" {
			s.openRowStart = len(s.rows)
		}
		s.rows = append(s.rows, entryRows(entry, width)...)
	}
}
