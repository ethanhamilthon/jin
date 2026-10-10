package ui

import (
	"errors"

	"jin/internal/core"
	"jin/internal/session"
)

type titleResult struct {
	id, title string
	err       error
	// manual is set for /title, whose new name is shown in the chat.
	manual bool
}

// nameNow names the focused session now, in the background, as the web
// "Generate title" item does.
func (a *app) nameNow() {
	s := a.active
	if err := a.noEnabledProvider(s); err != nil {
		a.report(err)
		return
	}
	if s.providerMissing {
		a.report(errors.New(missingProviderText(s.provider)))
		return
	}
	if !s.persisted {
		a.report(errors.New("Nothing to title yet: this session has no messages"))
		return
	}
	cfg, err := a.store.LoadConfig()
	if err != nil {
		a.report(err)
		return
	}
	id, providerID, model := s.id, s.provider, s.model
	go func() {
		title, err := session.NameSession(a.ctx, a.store, cfg, a.registry.SchemaJSON(), id, providerID, model)
		select {
		case a.titles <- titleResult{id: id, title: title, err: err, manual: true}:
		case <-a.ctx.Done():
		}
	}()
}

// autoTitle names the session once its user messages reach the configured
// count. The model runs in the background; a failure is shown once.
func (a *app) autoTitle(s *chatSession) {
	if a.store == nil || a.titles == nil || !s.persisted || s.providerMissing {
		return
	}
	cfg, err := a.store.LoadConfig()
	if err != nil || cfg.Title.After == 0 {
		return
	}
	messages, err := a.store.LoadMessages(s.id)
	if err != nil {
		return
	}
	turns := session.UserTurns(messages)
	if !session.TitleDue(cfg.Title, turns, s.titledAt) {
		return
	}
	s.titledAt = turns
	id, providerID, model := s.id, s.provider, s.model
	go func() {
		title, err := session.NameSession(a.ctx, a.store, cfg, a.registry.SchemaJSON(), id, providerID, model)
		select {
		case a.titles <- titleResult{id: id, title: title, err: err}:
		case <-a.ctx.Done():
		}
	}()
}

func (a *app) receiveTitle(r titleResult) {
	s := a.sessions[r.id]
	if s == nil {
		return
	}
	if r.err != nil {
		s.closeOpenEntry()
		s.appendEntry(chatEntry{kind: core.UpdateError, text: "Title was not generated: " + r.err.Error()})
		return
	}
	s.title = r.title
	if r.manual {
		s.closeOpenEntry()
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Session named: " + r.title})
	}
}
