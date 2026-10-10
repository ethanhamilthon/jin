package ui

import "jin/internal/tools"

// receiveBackendConfig reloads the settings a config event announces: the
// daemon changed the providers, the tools or the active model, and every
// client follows.
func (a *app) receiveBackendConfig() {
	cfg, err := a.store.LoadConfig()
	if err != nil {
		return
	}
	a.cfg = cfg
	if a.client != nil {
		a.client.Configure(cfg.Provider)
	}
	a.markMissingProviders()
	for _, s := range a.sessions {
		s.toolNames = tools.Without(cfg.ToolsDisabled)
	}
}

// receiveBackendTasks recounts the running background tasks of every session.
func (a *app) receiveBackendTasks() {
	counts := map[string]int{}
	for _, task := range a.taskList() {
		if task.Status == "running" {
			counts[task.Owner]++
		}
	}
	a.tasksRunning = counts
}
