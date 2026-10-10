package ui

import "jin/internal/store"

func (a *app) cycleChoice() {
	s := a.active
	if len(s.modelChoices) == 0 {
		return
	}
	current := -1
	for i, choice := range s.modelChoices {
		if choice.Provider == s.provider && choice.ID == s.model {
			current = i
			break
		}
	}
	next := s.modelChoices[(current+1)%len(s.modelChoices)]
	if next.Provider == s.provider && next.ID == s.model {
		return
	}
	effort := fallbackEffort
	if value, ok := a.cfg.ModelEfforts[store.EffortKey(next.Provider, next.ID)]; ok {
		effort = value
	} else if value, ok := a.cfg.ModelEfforts[next.ID]; ok {
		effort = value
	}
	if err := a.chooseModelSource(s, next.Provider, next.ID, effort); err != nil {
		s.persistenceError("Model was not changed", err)
	}
}
