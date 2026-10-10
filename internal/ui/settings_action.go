package ui

import (
	"context"
	"time"
)

func (a *app) settingsAction(title string, action func(context.Context) error, done func()) {
	ctx, cancel := context.WithTimeout(a.ctx, 10*time.Minute)
	sel := a.openList(title, nil, "", func(string) error { done(); return nil })
	sel.loading, sel.advance = true, true
	sel.onCancel = cancel
	go func() {
		defer cancel()
		err := action(ctx)
		options := []option{{label: "Done", value: "done"}}
		select {
		case a.loads <- loadResult{sel: sel, options: options, err: err}:
		case <-a.ctx.Done():
		}
	}()
}
