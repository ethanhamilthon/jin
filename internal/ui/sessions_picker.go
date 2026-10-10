package ui

import (
	"errors"
	"time"

	"jin/internal/store"
)

// newRow is the value of the "New session" row, which no session has.
const newRow = ""

// openProjectSessions is the second step of the picker: "New session", then
// the sessions of one project under their recency headers. Esc and Left go
// back to the projects.
func (a *app) openProjectSessions(path string) *selector {
	records, err := a.store.ListByPath(path)
	a.unread, _ = a.store.UnreadSessions(path)
	options := []option{{label: "New session", value: newRow}}
	for _, group := range groupSessions(records, a.runningSessions(records), time.Now()) {
		options = append(options, option{label: group.name, header: true})
		for _, rec := range group.rows {
			options = append(options, sessionOption(rec))
		}
	}
	current := ""
	if a.active != nil {
		current = a.active.id
	}
	sel := a.openList("Sessions · "+shortPath(path), options, current, func(id string) error {
		if id == newRow {
			return a.newProjectSession(path)
		}
		rec, ok, err := a.store.GetSession(id)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("session not found")
		}
		return a.resumeProjectSession(rec)
	})
	sel.twoLines, sel.leftBack = true, true
	sel.hint = "Enter open · Esc back"
	sel.onCancel = func() { a.openProjectList(path) }
	if err != nil {
		sel.err = err.Error()
	}
	return sel
}

func sessionOption(rec store.Session) option {
	title := rec.Title
	if title == "" {
		title = "Untitled"
	}
	detail := relativeTime(rec.UpdatedAt) + " · " + rec.Model + " · " + usageLine(rec.Usage)
	return option{label: title, detail: detail, value: rec.ID}
}

// runningSessions finds the sessions that work now: open in this process and
// answering, or owned by a live jin process according to the store.
func (a *app) runningSessions(records []store.Session) map[string]bool {
	running := map[string]bool{}
	for _, rec := range records {
		if s := a.sessions[rec.ID]; s != nil && (s.working || s.bash != nil && s.bash.running) {
			running[rec.ID] = true
		} else if _, alive, err := a.store.SessionOwner(rec.ID); err == nil && alive {
			running[rec.ID] = true
		}
	}
	return running
}
