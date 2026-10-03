package ui

import (
	"errors"
	"fmt"

	"jin/internal/editor"
)

// chooseEditor lists the supported editors; then runs once the choice is saved.
func (a *app) chooseEditor(then func() error) {
	options := make([]option, len(editor.Choices))
	for i, name := range editor.Choices {
		detail := "installed"
		if !editor.Installed(name) {
			detail = "not installed"
		}
		options[i] = option{label: name, detail: detail, value: name}
	}
	a.openList("Editor", options, a.cfg.Editor, func(name string) error {
		if !editor.Valid(name) {
			return errors.New("unknown editor")
		}
		if !editor.Installed(name) {
			return fmt.Errorf("%s is not installed", name)
		}
		if err := a.store.SaveEditor(name); err != nil {
			return err
		}
		a.cfg.Editor = name
		return then()
	})
}

// editFile opens path in the chosen editor, asking for one first if needed,
// and calls done once the editor exits cleanly.
func (a *app) editFile(path string, done func() error) error {
	if a.cfg.Editor == "" {
		a.chooseEditor(func() error { return a.editFile(path, done) })
		return nil
	}
	if err := a.runEditor(path); err != nil {
		return err
	}
	return done()
}

// runEditor hands the terminal to the editor and takes it back afterwards.
func (a *app) runEditor(path string) error {
	if !editor.Installed(a.cfg.Editor) {
		return fmt.Errorf("%s is not installed", a.cfg.Editor)
	}
	if err := a.runExternal(editor.Command(a.cfg.Editor, path)); err != nil {
		return fmt.Errorf("%s: %w", a.cfg.Editor, err)
	}
	return nil
}
