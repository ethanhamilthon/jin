package ui

import "jin/internal/store"

// newProjectSession starts a fresh session in dir and focuses it.
func (a *app) newProjectSession(dir string) error {
	if err := a.store.RememberProject(dir, ""); err != nil {
		return err
	}
	a.focus(a.startSessionAt(dir, newSessionID(), a.cfg.ActiveProvider, a.cfg.Model, a.cfg.Effort, nil, a.introEntriesAt(dir)))
	a.askHooksTrust()
	return nil
}

func (a *app) resumeProjectSession(rec store.Session) error {
	if err := a.resumeSession(rec); err != nil {
		return err
	}
	a.askHooksTrust()
	return nil
}

func sameProject(a, b string) bool {
	if a == b {
		return true
	}
	left, err := projectPath(a, "")
	if err != nil {
		return false
	}
	right, err := projectPath(b, "")
	return err == nil && left == right
}
