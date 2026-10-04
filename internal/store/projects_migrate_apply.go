package store

import "database/sql"

func rebuildProjectRows(tx *sql.Tx, projects, saved []Project, sessions []migrationSession) error {
	groups := map[string]*migrationProject{}
	oldIDs := map[string]string{}
	groupFor := func(path string) (*migrationProject, string, error) {
		canonical, err := canonicalStoredPath(path)
		if err != nil {
			return nil, "", err
		}
		group := groups[canonical]
		if group == nil {
			group = &migrationProject{path: canonical}
			groups[canonical] = group
		}
		return group, canonical, nil
	}
	for _, project := range projects {
		group, key, err := groupFor(project.Path)
		if err != nil {
			return err
		}
		mergeProject(group, project)
		oldIDs[project.ID] = key
	}
	sessionKeys := make(map[string]string, len(sessions))
	for _, session := range sessions {
		var key string
		if session.path != "" {
			var err error
			_, key, err = groupFor(session.path)
			if err != nil {
				return err
			}
		} else {
			key = oldIDs[session.projectID]
		}
		if key == "" {
			continue
		}
		sessionKeys[session.id] = key
		group := groups[key]
		if session.created > 0 && (group.created == 0 || session.created < group.created) {
			group.created = session.created
		}
		if session.updated >= group.latestUpdate {
			group.latestUpdate = session.updated
			group.latestSession = session.id
		}
	}
	for _, project := range saved {
		if project.Path == "" {
			continue
		}
		group, _, err := groupFor(project.Path)
		if err != nil {
			return err
		}
		mergeProject(group, project)
	}
	return persistProjectGroups(tx, groups, sessions, sessionKeys)
}

func mergeProject(group *migrationProject, project Project) {
	if group.id == "" {
		group.id = project.ID
	}
	if group.name == "" {
		group.name = project.Name
	}
	if project.LastSession != "" {
		group.lastSessions = append(group.lastSessions, project.LastSession)
	}
	created, opened := project.CreatedAt.Unix(), project.LastOpenedAt.Unix()
	if project.CreatedAt.IsZero() {
		created = 0
	}
	if project.LastOpenedAt.IsZero() {
		opened = 0
	}
	if created > 0 && (group.created == 0 || created < group.created) {
		group.created = created
	}
	if opened > group.opened {
		group.opened = opened
	}
}
