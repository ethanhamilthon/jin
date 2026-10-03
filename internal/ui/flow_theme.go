package ui

// openThemeFlow lists the themes. Moving through the list shows each one at
// once; Enter keeps it, Esc goes back to the theme that was on.
func (a *app) openThemeFlow() {
	before := a.cfg.Theme
	options := make([]option, len(themes))
	for i, t := range themes {
		options[i] = option{label: t.name, value: t.name}
	}
	if before == "" {
		before = themes[0].name
	}
	sel := a.openList("Theme · ↑/↓ preview · Enter keep · Esc cancel", options, before, func(name string) error {
		if err := a.store.SaveTheme(name); err != nil {
			return err
		}
		a.cfg.Theme = name
		a.useTheme(name)
		return nil
	})
	sel.onMove = a.useTheme
	sel.onCancel = func() { a.useTheme(before) }
}

// useTheme switches the palette and redraws every session with it.
func (a *app) useTheme(name string) {
	applyTheme(themeByName(name))
	if a.screen != nil {
		a.screen.SetStyle(base)
	}
	for _, s := range a.sessions {
		s.rebuildRows(s.width)
	}
	if a.active != nil && a.sessions[a.active.id] != a.active {
		a.active.rebuildRows(a.active.width)
	}
}
