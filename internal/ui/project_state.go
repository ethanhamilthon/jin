package ui

import "sort"

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

func (a *app) watchAsyncPaths() {
	if a.asyncPaths == nil {
		return
	}
	seen := map[string]bool{}
	for _, s := range a.sessions {
		if s.path != "" {
			seen[s.path] = true
		}
	}
	paths := make([]string, 0, len(seen))
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	select {
	case a.asyncPaths <- paths:
	default:
		select {
		case <-a.asyncPaths:
		default:
		}
		a.asyncPaths <- paths
	}
}
