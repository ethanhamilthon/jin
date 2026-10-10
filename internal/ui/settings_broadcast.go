package ui

import "jin/internal/daemon"

// settingsChanged tells the daemon that a setting changed in the store. Every
// other client reloads it; this client already shows the change.
func (a *app) settingsChanged() {
	if a.backend == nil {
		return
	}
	_ = a.backend.Command(a.ctx, daemon.Command{Action: "settings-changed"}, nil)
}
