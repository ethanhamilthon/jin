package ui

import (
	"strings"

	"jin/internal/hooks"
	"jin/internal/store"
)

// projectHooksTrusted reports whether the hooks in .jin/hooks of the working
// directory may run. Until the user answers, they do not.
func (a *app) projectHooksTrusted() bool {
	if a.store == nil {
		return false
	}
	trust, err := a.store.HooksTrust(a.dir)
	return err == nil && trust == store.Trusted
}

// askHooksTrust asks once per directory whether its project hooks may run,
// because their {{commands}} run on the user's machine. It returns false
// when there is nothing to ask.
func (a *app) askHooksTrust() bool {
	names, _ := hooks.ListProject(a.dir)
	if len(names) == 0 || a.store == nil {
		return false
	}
	if trust, err := a.store.HooksTrust(a.dir); err != nil || trust != store.TrustUnknown {
		return false
	}
	options := []option{
		{label: "No, keep them off", value: "no"},
		{label: "Yes, I trust this repository", value: "yes"},
	}
	title := "Run the project hooks of this folder? .jin/hooks: " + strings.Join(names, ", ") +
		" · their {{commands}} run on your machine"
	a.openList(title, options, "no", func(answer string) error {
		if err := a.store.SaveHooksTrust(a.dir, answer == "yes"); err != nil {
			return err
		}
		if answer == "yes" {
			a.restartBlank()
		}
		return nil
	})
	return true
}

// restartBlank replaces the focused session, when nothing was sent to it, so
// it starts again with the hooks the user just allowed.
func (a *app) restartBlank() {
	if s := a.active; s != nil && !s.persisted && len(s.pending) == 0 && !s.working {
		draft, cursor := s.input, s.cursor
		a.newSession()
		a.active.input, a.active.cursor = draft, cursor
	}
}
