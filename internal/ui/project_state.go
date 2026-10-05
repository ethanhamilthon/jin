package ui

func (a *app) rememberFocus(s *chatSession) {
	if a.lastProjectSession == nil {
		a.lastProjectSession = map[string]string{}
	}
	if s.path == "" {
		return
	}
	a.lastProjectSession[s.path] = s.id
	if a.store != nil {
		id := ""
		if s.persisted {
			id = s.id
		}
		if err := a.store.RememberProject(s.path, id); err != nil {
			s.persistenceError("Project was not saved", err)
		}
	}
}
