package ui

import (
	"os"
	"strings"

	"jin/internal/hooks"
)

// toggleHook switches a hook on or off. A project hook of a folder that is
// not trusted asks for trust instead.
func (a *app) toggleHook(sel *selector, value string) {
	if value == "" {
		return
	}
	if strings.HasPrefix(value, projectPrefix) && !a.projectHooksTrusted() {
		a.confirmTrust(value)
		return
	}
	disabled := hooks.Toggle(a.cfg.HooksDisabled, a.hookKey(value))
	if err := a.store.SaveHooksDisabled(disabled); err != nil {
		sel.err = err.Error()
		return
	}
	a.cfg.HooksDisabled = disabled
}

func (a *app) confirmTrust(value string) {
	options := []option{{label: "No", value: "no"}, {label: "Yes, run the hooks of this folder", value: "yes"}}
	a.openList("Trust .jin/hooks of this folder? Their {{commands}} run on your machine", options, "no", func(answer string) error {
		if answer == "yes" {
			if err := a.store.SaveHooksTrust(a.dir, true); err != nil {
				return err
			}
		}
		a.showHooks(value)
		return nil
	})
}

func (a *app) deleteHook(value string) error {
	path, err := a.hookPath(value)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	disabled := hooks.Forget(a.cfg.HooksDisabled, a.hookKey(value))
	if err := a.store.SaveHooksDisabled(disabled); err != nil {
		return err
	}
	a.cfg.HooksDisabled = disabled
	return nil
}
