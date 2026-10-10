package session

// UseSharedQueue keeps waiting messages in the manager until the current turn ends.
func (m *Manager) UseSharedQueue() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sharedQueue = true
}
