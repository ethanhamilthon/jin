package ui

import (
	"errors"
	"maps"

	"jin/internal/search"
)

func (a *app) openSearchFlow() {
	current := a.cfg.Search
	var options []option
	for _, backend := range search.Backends {
		detail := "no key needed"
		if backend.NeedsKey() {
			detail = "API key missing"
			if current.Keys[backend] != "" {
				detail = "API key set"
			}
		}
		options = append(options, option{label: string(backend), detail: detail, value: string(backend)})
	}
	a.openList("Web search backend", options, string(current.Backend), func(value string) error {
		backend := search.Backend(value)
		if !backend.NeedsKey() {
			return a.saveSearch(backend, "")
		}
		a.openField(value+" API key", current.Keys[backend], true, func(key string) error {
			if key == "" {
				return errors.New("API key is required")
			}
			return a.saveSearch(backend, key)
		})
		return nil
	})
}

func (a *app) saveSearch(backend search.Backend, key string) error {
	cfg := search.Config{Backend: backend, Keys: maps.Clone(a.cfg.Search.Keys)}
	if key != "" {
		cfg.Keys[backend] = key
	}
	if err := a.store.SaveSearch(cfg); err != nil {
		return err
	}
	a.cfg.Search = cfg
	a.search.Set(cfg)
	a.refreshIntro()
	return nil
}

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
