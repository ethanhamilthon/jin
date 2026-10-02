package ui

// openJinDocsFlow switches the jin docs pointer on or off. It reaches new
// sessions only: the system prompt is built when a session starts.
func (a *app) openJinDocsFlow() {
	options := []option{
		{label: "Jin docs in context", value: "docs", choices: []string{"On", "Off"}, chosen: indexOf(!a.cfg.JinDocs)},
	}
	sel := a.openList("Jin docs · ←/→ change · applies to new sessions", options, "", func(string) error { return nil })
	sel.keepOpen = true
	sel.onChoice = func(_ string, chosen int) error {
		if err := a.store.SaveJinDocs(chosen == 0); err != nil {
			return err
		}
		a.cfg.JinDocs = chosen == 0
		return nil
	}
}
