package ui

import (
	"errors"
	"fmt"
	"os/exec"

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

// runExternal suspends the TUI, runs cmd on the real terminal and resumes.
func (a *app) runExternal(cmd *exec.Cmd) error {
	if err := a.screen.Suspend(); err != nil {
		return err
	}
	runErr := cmd.Run()
	resumeErr := a.screen.Resume()
	a.screen.Sync()
	if runErr != nil {
		return runErr
	}
	return resumeErr
}

// runTUI is /tui: a full-screen program on the real terminal.
// The agent keeps running while the program is open: updates are still applied,
// so the notification sound plays the moment the agent finishes.
func (a *app) runTUI(command string) {
	if err := a.screen.Suspend(); err != nil {
		a.report(fmt.Errorf("%s: %w", command, err))
		return
	}
	wasBlurred := a.blurred
	a.blurred = true // the program owns the terminal, so jin counts as not focused
	cmd := editor.Shell(command, a.dir)
	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()
	var runErr error
wait:
	for {
		select {
		case runErr = <-done:
			break wait
		case tagged := <-a.updates:
			a.applyUpdate(tagged.id, tagged.update)
		case result := <-a.loads:
			a.receiveLoad(result)
		case result := <-a.bashDone:
			a.receiveBash(result)
		case result := <-a.modelsLoaded:
			a.receiveModels(result)
		}
		a.flushPending()
	}
	a.blurred = wasBlurred
	resumeErr := a.screen.Resume()
	a.screen.Sync()
	if runErr == nil {
		runErr = resumeErr
	}
	if runErr != nil {
		a.report(fmt.Errorf("%s: %w", command, runErr))
	}
}
