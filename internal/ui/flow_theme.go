package ui

import "strings"

// openThemeFlow lists the built-in and custom themes. Moving through the list
// shows each one at once; Enter keeps it, Esc goes back to the theme that was on.
func (a *app) openThemeFlow() {
	before := a.cfg.Theme
	if before == "" {
		before = themes[0].name
	}
	a.showThemes(before, before)
}

func (a *app) showThemes(current, before string) {
	all := allThemes()
	options := make([]option, len(all))
	for i, t := range all {
		options[i] = option{label: t.name, value: t.name}
		if i >= len(themes) {
			options[i].detail = "custom"
		}
	}
	sel := a.openList("Theme", options, current, func(name string) error {
		if err := a.store.SaveTheme(name); err != nil {
			return err
		}
		a.cfg.Theme = name
		a.useTheme(name)
		a.settingsChanged()
		return nil
	})
	sel.hint = "/ search · r reload · Enter keep"
	sel.onMove = a.useTheme
	sel.onCancel = func() { a.useTheme(before) }
	sel.actions = map[rune]func(string){
		'r': func(name string) { a.showThemes(name, before) },
	}
	if _, errs := customThemes(); len(errs) > 0 {
		sel.err = "theme files: " + strings.Join(errs, "; ")
	}
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
