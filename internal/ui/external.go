package ui

import (
	"fmt"
	"jin/internal/editor"
	"os/exec"
)

// runExternal suspends the TUI, runs cmd on the real terminal and resumes.
// The agents keep working while the program is open: updates, background
// task results, queued messages and the notification sound are still handled,
// as if the screen were there. Without this the event loop would stand still
// for as long as an editor is open, and a full update channel would stall
// the agents.
func (a *app) runExternal(cmd *exec.Cmd) error {
	if err := a.screen.Suspend(); err != nil {
		return err
	}
	wasBlurred := a.blurred
	a.blurred = true // the program owns the terminal, so jin counts as not focused
	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()
	runErr := a.serveUntil(done)
	a.blurred = wasBlurred
	resumeErr := a.screen.Resume()
	a.screen.Sync()
	if runErr != nil {
		return runErr
	}
	return resumeErr
}

// serveUntil handles everything the background sources deliver until done
// reports, and returns what done reported.
func (a *app) serveUntil(done <-chan error) error {
	for {
		select {
		case err := <-done:
			return err
		case tagged := <-a.updates:
			a.applyUpdate(tagged.id, tagged.update)
		case result := <-a.loads:
			a.receiveLoad(result)
		case result := <-a.bashDone:
			a.receiveBash(result)
		case result := <-a.modelsLoaded:
			a.receiveModels(result)
		case batch := <-a.asyncs:
			a.receiveAsync(batch)
		case ev := <-a.rendered:
			a.receiveRender(ev)
		}
		a.flushPending()
	}
}

// runTUI is /tui: a full-screen program on the real terminal.
func (a *app) runTUI(command string) {
	if err := a.runExternal(editor.Shell(command, a.dir)); err != nil {
		a.report(fmt.Errorf("%s: %w", command, err))
	}
}
