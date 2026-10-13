package session

// dropQueued forgets every request that no turn has taken yet.
func (s *Session) dropQueued() {
	s.pending = nil
	for {
		select {
		case request := <-s.requests:
			s.agent.Discard(request)
			s.inflight--
		default:
			return
		}
	}
}
