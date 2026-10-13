package ui

import (
	"jin/internal/core"
	"jin/internal/session"
)

type titleResult struct {
	id, title string
	err       error
}

func (s *chatSession) startTitle(prompt string) {
	if s.app == nil {
		return
	}
	id, providerID, model, effort := s.id, s.provider, s.model, s.effort
	go func() {
		cfg, err := s.store.LoadConfig()
		var title string
		if err == nil {
			title, err = session.NamePrompt(s.app.ctx, s.store, cfg, id, providerID, model, effort, prompt)
		}
		select {
		case s.app.titles <- titleResult{id: id, title: title, err: err}:
		case <-s.app.ctx.Done():
		}
	}()
}

func (a *app) receiveTitle(result titleResult) {
	s := a.sessions[result.id]
	if s == nil {
		return
	}
	if result.err != nil {
		s.appendEntry(chatEntry{kind: core.UpdateError, text: "Title was not generated: " + result.err.Error()})
		return
	}
	if result.title != "" {
		s.title = result.title
	}
}
