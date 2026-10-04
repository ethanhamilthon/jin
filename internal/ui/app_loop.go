package ui

import (
	"context"
	"time"

	"jin/internal/pricing"
)

// loop draws the screen and handles one event at a time until the user
// quits or ctx ends.
func (a *app) loop(ctx context.Context, prices <-chan pricing.Table) error {
	a.newSession()
	if entry, ok := a.whatsNew(); ok {
		a.active.appendEntry(entry)
	}
	if !a.askHooksTrust() {
		a.startOnboarding()
	}
	go a.pollAsync()
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()
	for !a.quit {
		a.draw()
		select {
		case <-ctx.Done():
			return nil
		case tagged := <-a.updates:
			a.applyUpdate(tagged.id, tagged.update)
		case result := <-a.loads:
			a.receiveLoad(result)
		case result := <-a.bashDone:
			a.receiveBash(result)
		case result := <-a.voiceDone:
			a.receiveVoice(result)
		case result := <-a.modelsLoaded:
			a.receiveModels(result)
		case batch := <-a.asyncs:
			a.receiveAsync(batch)
		case ev := <-a.rendered:
			a.receiveRender(ev)
		case tag := <-a.newVersion:
			a.receiveUpdate(tag)
		case table := <-prices:
			a.setPricing(table)
		case <-ticker.C:
			a.tick()
			a.voiceTick()
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
