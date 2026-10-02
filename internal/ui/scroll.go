package ui

func (s *chatSession) scrollBy(delta int) {
	s.scroll = max(0, s.scroll+delta)
}
