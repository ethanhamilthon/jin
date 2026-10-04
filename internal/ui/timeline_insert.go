package ui

func (s *chatSession) insertBeforeStream(entry chatEntry) {
	s.flushStream()
	index := len(s.history)
	if s.openKind != "" && index > 0 {
		index--
	}
	s.history = append(s.history, chatEntry{})
	copy(s.history[index+1:], s.history[index:])
	s.history[index] = entry
	for i := range s.stream.attempt {
		if s.stream.attempt[i] >= index {
			s.stream.attempt[i]++
		}
	}
	oldRows := len(s.rows)
	s.rebuildRows(s.width)
	if s.scroll > 0 {
		s.scroll += len(s.rows) - oldRows
	}
}
