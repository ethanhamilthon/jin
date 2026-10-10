package web

// dropDevice closes every open page of the device whose access ended. The
// guard refuses its cookie from then on, so the pages cannot reconnect.
func (h *hub) dropDevice(id string) {
	if id == "" {
		return
	}
	h.kick(id)
}
