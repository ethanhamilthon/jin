package ui

import (
	"context"
	"github.com/gdamore/tcell/v3"
	"jin/internal/tasks"
)

// Run shows the TUI until the user quits. A returned DataAction must be
// carried out after the database is closed.
func Run(ctx context.Context, deps Deps) (*DataAction, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}
	defer screen.Fini()
	detectColors(screen)
	applyTheme(themeByName(deps.Config.Theme))
	screen.SetStyle(base)
	screen.EnableMouse(tcell.MouseDragEvents)
	screen.EnablePaste()
	screen.EnableFocus()
	screen.SetCursorStyle(tcell.CursorStyleSteadyBar)
	w, _ := screen.Size()
	a := &app{
		screen: screen, ctx: ctx, cancel: cancel, store: deps.Store, cfg: deps.Config, dir: deps.Dir, version: deps.Version,
		client: deps.Client, registry: deps.Registry, width: w, fold: foldMode(deps.Config.Fold),
		lastProjectSession: map[string]string{}, taskEvents: tasks.Shared().Events(),
		sessions: map[string]*chatSession{}, updates: make(chan taggedUpdate, 256), loads: make(chan loadResult, 4), modelsLoaded: make(chan modelsResult, 1), bashDone: make(chan bashResult, 4), rendered: make(chan renderEvent, 32), tasksRunning: map[string]int{}, newVersion: make(chan string, 1), titles: make(chan titleResult, 4),
	}
	defer a.shutdownSessions()
	err = a.loop(ctx, deps.Pricing)
	return a.dataAction, err
}
