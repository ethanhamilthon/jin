package ui

import (
	"jin/internal/pricing"
)

func (a *app) setPricing(table pricing.Table) {
	a.pricing = table
	for _, s := range a.sessions {
		s.pricing = table
	}
}

func (a *app) focus(s *chatSession) {
	if s == nil {
		return
	}
	changed, wasUnread := a.active != s, s.unread
	if changed {
		a.slash, a.mention, a.file = nil, nil, nil
	}
	if a.panes != nil {
		if visible := a.paneForSession(s); visible != nil {
			a.focused = visible
		} else {
			leaf := a.focusedLeaf()
			if leaf == nil {
				leaf = firstPaneLeaf(a.panes)
			}
			if leaf == nil {
				a.panes = &paneNode{session: s}
				leaf = a.panes
			} else {
				leaf.session = s
			}
			a.focused = leaf
		}
	}
	a.active = s
	if s.path != "" {
		a.dir = s.path
	}
	if a.panes == nil {
		a.focused = nil
	}
	if changed {
		a.rememberFocus(s)
	}
	s.unread = false
	if s.persisted && s.store != nil && (changed || wasUnread) {
		_ = s.store.SetUnread(s.id, false)
	}
}
