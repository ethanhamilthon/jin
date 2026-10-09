package web

var devicesEvent = map[string]string{"type": "devices"}

// online lists the devices that have a page open right now.
func (h *hub) online() map[string]bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	ids := map[string]bool{}
	for _, id := range h.clients {
		ids[id] = true
	}
	return ids
}

// kick closes every open page of a device whose access was revoked.
func (h *hub) kick(id string) {
	h.mu.Lock()
	for c, device := range h.clients {
		if device == id {
			delete(h.clients, c)
			close(c)
		}
	}
	h.mu.Unlock()
	h.publish(devicesEvent)
}
