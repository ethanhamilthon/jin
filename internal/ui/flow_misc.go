package ui

import "errors"

func (a *app) openSessionsFlow() *selector {
	records, err := a.store.ListByPath(a.dir)
	a.unread, _ = a.store.UnreadSessions(a.dir)
	options := make([]option, 0, len(records))
	for _, rec := range records {
		detail := relativeTime(rec.UpdatedAt) + " · " + rec.Model + " · " + usageLine(rec.Usage)
		options = append(options, option{label: rec.Title, detail: detail, value: rec.ID})
	}
	sel := a.openList("Sessions · "+shortPath(a.dir), options, a.active.id, func(id string) error {
		for _, rec := range records {
			if rec.ID == id {
				return a.resumeSession(rec)
			}
		}
		return errors.New("session not found")
	})
	sel.twoLines = true
	if err != nil {
		sel.err = err.Error()
	}
	return sel
}

func (a *app) requestQuit() {
	if !a.anyWorking() {
		a.quit = true
		return
	}
	options := []option{{label: "Quit", detail: "stop running requests", value: "quit"}, {label: "Cancel", value: "cancel"}}
	a.openList("A request is still running. Quit jin?", options, "cancel", func(value string) error {
		a.quit = value == "quit"
		return nil
	})
}
