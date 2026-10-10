package session

// flush hands queued requests to each agent without blocking.
func (m *Manager) flush() {
	for _, s := range m.sessions {
		if !s.ready || s.paused || (m.sharedQueue && s.inflight > 0) {
			continue
		}
		moved := false
	deliver:
		for len(s.pending) > 0 {
			select {
			case s.prompts <- s.pending[0]:
				s.inflight++
				if s.pending[0].Interactive {
					s.agent.DetachTools()
				}
				s.pending = s.pending[1:]
				moved = true
				if m.sharedQueue {
					break deliver
				}
			default:
				break deliver
			}
		}
		if moved {
			s.emitState()
		}
	}
}
