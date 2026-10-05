package ui

import "github.com/gdamore/tcell/v3"

// projectActivity is what the store knows about sessions that are not open:
// the project of each session and the unread ones.
type projectActivity struct {
	projects map[string]string
	sessions map[string][]string
	unread   map[string]bool
}

func (a *app) loadProjectActivity() projectActivity {
	projects, _ := a.store.SessionProjects()
	unread, _ := a.store.AllUnread()
	sessions := map[string][]string{}
	for id, path := range projects {
		sessions[path] = append(sessions[path], id)
	}
	return projectActivity{projects: projects, sessions: sessions, unread: unread}
}

// projectMark flags a project. The open one is always green. For the others,
// in order of priority: a blinking blue dot while one of its sessions
// answers, a steady blue dot for an unread answer, a blinking purple dot
// while a background task runs.
func (a *app) projectMark(path string, act projectActivity) (string, tcell.Style) {
	if path == a.dir {
		return "●", dotGreen
	}
	var working, unread, async bool
	for _, id := range act.sessions[path] {
		unread = unread || act.unread[id] && a.sessions[id] == nil
		async = async || a.tasksRunning[id] > 0
	}
	for id, s := range a.sessions {
		if s.path != path && act.projects[id] != path {
			continue
		}
		working = working || s.working || len(s.pending) > 0
		unread = unread || s.unread
		async = async || a.tasksRunning[id] > 0
	}
	blink := a.frame%8 < 5
	switch {
	case working && blink:
		return "●", dotBlue
	case working:
		return " ", dotBlue
	case unread:
		return "●", dotBlue
	case async && blink:
		return "●", dotPurple
	}
	return " ", dotPurple
}
