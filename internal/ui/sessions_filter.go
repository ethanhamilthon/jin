package ui

import (
	"errors"
	"strings"
)

func isSessionsSelector(sel *selector) bool {
	if sel == nil {
		return false
	}
	return strings.HasPrefix(sel.title, "Sessions · ")
}

func (a *app) initSessionsFilter(sel *selector) {
	if !isSessionsSelector(sel) {
		return
	}
	sel.allOptions = append([]option(nil), sel.options...)
	if a.store == nil {
		return
	}
	origMap := make(map[string]option, len(sel.options))
	for _, opt := range sel.options {
		origMap[opt.value] = opt
	}
	sel.filter = func(words []string) ([]option, error) {
		results, err := a.store.SearchSessions(a.dir, false, words)
		if err != nil {
			return nil, err
		}
		opts := make([]option, 0, len(results))
		for _, r := range results {
			detail := r.Snippet
			if orig, ok := origMap[r.ID]; ok && (detail == "" || detail == r.Title) {
				detail = orig.detail
			} else if detail == "" {
				detail = r.Date
			}
			title := r.Title
			if title == "" {
				if orig, ok := origMap[r.ID]; ok && orig.label != "" {
					title = orig.label
				} else {
					title = r.ID
				}
			}
			opts = append(opts, option{label: title, detail: detail, value: r.ID})
		}
		return opts, nil
	}
	origSubmit := sel.submit
	sel.submit = func(id string) error {
		if origSubmit != nil && origSubmit(id) == nil {
			return nil
		}
		rec, ok, err := a.store.GetSession(id)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("session not found")
		}
		return a.resumeSession(rec)
	}
}

func (a *app) drawFilterHeader(y, w int, sel *selector, minX int) {
	if sel == nil || sel.filter == nil {
		return
	}
	f := strings.TrimSpace(sel.filterText())
	if f == "" {
		return
	}
	label := " /" + f + " "
	if startX := w - len([]rune(label)) - 2; startX >= minX {
		put(a.screen, startX, y, label, accent.Bold(true))
		return
	}
	if avail := w - minX - 6; avail > 0 {
		label = " /" + truncate(f, avail) + " "
		put(a.screen, w-len([]rune(label))-2, y, label, accent.Bold(true))
	}
}
