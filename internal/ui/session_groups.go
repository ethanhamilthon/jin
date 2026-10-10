package ui

import (
	"slices"
	"time"

	"jin/internal/store"
)

// groupNames are the headers of the session step, in the order they show.
var groupNames = []string{"Running", "Last hour", "Last 6 hours", "Today", "This Week", "Older"}

// sessionGroup is one header of the session step and its sessions, in the
// order the store lists them.
type sessionGroup struct {
	name string
	rows []store.Session
}

// groupOf names the first group a session matches.
func groupOf(updated time.Time, running bool, now time.Time) string {
	age := now.Sub(updated)
	switch {
	case running:
		return "Running"
	case age < time.Hour:
		return "Last hour"
	case age < 6*time.Hour:
		return "Last 6 hours"
	case !updated.Before(startOfDay(now)):
		return "Today"
	case age < 7*24*time.Hour:
		return "This Week"
	}
	return "Older"
}

func startOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// groupSessions files the sessions under their groups and drops the empty ones.
func groupSessions(rows []store.Session, running map[string]bool, now time.Time) []sessionGroup {
	groups := make([]sessionGroup, len(groupNames))
	for i, name := range groupNames {
		groups[i].name = name
	}
	for _, rec := range rows {
		i := slices.Index(groupNames, groupOf(rec.UpdatedAt, running[rec.ID], now))
		groups[i].rows = append(groups[i].rows, rec)
	}
	return slices.DeleteFunc(groups, func(g sessionGroup) bool { return len(g.rows) == 0 })
}
