package ui

import (
	"context"
	"errors"
	"testing"
	"time"
)

func runAction(t *testing.T, a *app, err error) (continued bool) {
	t.Helper()
	a.ctx, a.loads = context.Background(), make(chan loadResult, 1)
	a.settingsAction("Work", func(context.Context) error { return err }, func() { continued = true })
	select {
	case result := <-a.loads:
		a.receiveLoad(result)
	case <-time.After(5 * time.Second):
		t.Fatal("action did not finish")
	}
	return continued
}

func TestSettingsActionContinuesAloneOnSuccessAndStopsOnFailure(t *testing.T) {
	a, _ := layoutApp(t)
	if !runAction(t, a, nil) {
		t.Fatal("a finished action must continue without a keypress")
	}
	b, _ := layoutApp(t)
	if runAction(t, b, errors.New("boom")) || b.sel == nil || b.sel.err != "boom" {
		t.Fatalf("a failed action must stay open with its error: %+v", b.sel)
	}
}
