package ui

import (
	"context"
	"errors"
	"time"

	"jin/internal/pricing"
)

// loop draws the screen and handles one event at a time until the user
// quits or ctx ends.
func (a *app) loop(ctx context.Context, prices <-chan pricing.Table) error {
	dir, err := projectPath(a.dir, "")
	if err != nil {
		return err
	}
	a.dir = dir
	if _, err := a.store.EnsureProject(dir); err != nil {
		return err
	}
	a.newSession()
	if a.active == nil {
		return errors.New("could not create a session through the daemon")
	}
	a.initPanes()
	if entry, ok := a.whatsNew(); ok {
		a.active.appendEntry(entry)
	}
	if !a.askHooksTrust() {
		a.startOnboarding()
	}
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()
	for !a.quit {
		a.draw()
		select {
		case <-ctx.Done():
			return nil
		case event, open := <-a.backendEvents:
			if !open {
				a.backendEvents = nil
			} else {
				a.receiveBackend(event)
			}
		case tagged := <-a.updates:
			a.applyUpdate(tagged.id, tagged.update)
		case result := <-a.loads:
			a.receiveLoad(result)
		case result := <-a.bashDone:
			a.receiveBash(result)
		case result := <-a.modelsLoaded:
			a.receiveModels(result)
		case event := <-a.taskEvents:
			a.receiveTask(event)
		case ev := <-a.rendered:
			a.receiveRender(ev)
		case result := <-a.titles:
			a.receiveTitle(result)
		case tag := <-a.newVersion:
			a.receiveUpdate(tag)
		case table := <-prices:
			a.setPricing(table)
		case <-ticker.C:
			a.tick()
		case event, ok := <-a.screen.EventQ():
			if !ok {
				return nil
			}
			a.handleEvent(event)
		}
		a.flushPending()
	}
	return nil
}
