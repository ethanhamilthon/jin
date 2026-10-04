package ui

import "jin/internal/store"

func (a *app) switchProject(path string) error {
	dir, err := projectPath(path, a.dir)
	if err != nil {
		return err
	}
	projects, err := a.store.Projects()
	if err != nil {
		return err
	}
	last := ""
	for _, project := range projects {
		if sameProject(project.Path, dir) {
			last = project.LastSession
		}
	}
	if id := a.lastProjectSession[dir]; id != "" {
		last = id
	}
	if s := a.sessions[last]; s != nil && sameProject(s.path, dir) {
		a.focus(s)
		return nil
	}
	if last != "" {
		rec, found, err := a.store.GetSession(last)
		if err != nil {
			return err
		}
		if found && sameProject(rec.Path, dir) {
			return a.resumeProjectSession(rec)
		}
	}
	records, err := a.store.ListByPath(dir)
	if err != nil {
		return err
	}
	if len(records) > 0 {
		return a.resumeProjectSession(records[0])
	}
	if err := a.store.RememberProject(dir, ""); err != nil {
		return err
	}
	s := a.startSessionAt(dir, newSessionID(), a.cfg.ActiveProvider, a.cfg.Model, a.cfg.Effort, nil, a.introEntriesAt(dir))
	a.focus(s)
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
