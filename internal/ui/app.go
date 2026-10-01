package ui

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v3"

	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/search"
	"jin/internal/store"
	"jin/internal/tools"
)

type mode uint8

const (
	modeNormal mode = iota
	modeInsert
)

type Deps struct {
	Store    *store.DB
	Config   store.Config
	Client   *provider.Client
	Search   *search.Settings
	Registry *tools.Registry
	Pricing  <-chan pricing.Table
	Dir      string
}

type app struct {
	screen   tcell.Screen
	ctx      context.Context
	store    *store.DB
	cfg      store.Config
	dir      string
	client   *provider.Client
	search   *search.Settings
	registry *tools.Registry
	pricing  pricing.Table
	sessions map[string]*chatSession
	active   *chatSession
	updates  chan taggedUpdate
	loads    chan loadResult
	mode     mode
	sel      *selector
	unread   map[string]bool
	pasting  bool
	frame    int
	width    int
	quit     bool
}

func Run(ctx context.Context, deps Deps) error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := screen.Init(); err != nil {
		return err
	}
	defer screen.Fini()
	screen.SetStyle(base)
	screen.EnableMouse(tcell.MouseDragEvents)
	screen.EnablePaste()
	w, _ := screen.Size()
	a := &app{
		screen: screen, ctx: ctx, store: deps.Store, cfg: deps.Config, dir: deps.Dir,
		client: deps.Client, search: deps.Search, registry: deps.Registry, width: w,
		sessions: map[string]*chatSession{}, updates: make(chan taggedUpdate, 256), loads: make(chan loadResult, 4),
	}
	defer a.markInterruptedUnread()
	a.newSession()
	if a.cfg.Provider.Ready() {
		a.mode = modeInsert
	}
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
		case table := <-deps.Pricing:
			a.setPricing(table)
		case <-ticker.C:
			a.frame++
		case event, ok := <-screen.EventQ():
			if !ok {
				return nil
			}
			a.handleEvent(event)
		}
		a.flushPending()
	}
	return nil
}
