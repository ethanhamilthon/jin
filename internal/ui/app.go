package ui

import (
	"context"

	"github.com/gdamore/tcell/v3"

	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
	"jin/internal/tools"
)

type Deps struct {
	Store    *store.DB
	Config   store.Config
	Client   *provider.Client
	Registry *tools.Registry
	Pricing  <-chan pricing.Table
	Dir      string
	Version  string
}

type app struct {
	screen    tcell.Screen
	ctx       context.Context
	store     *store.DB
	cfg       store.Config
	dir       string
	version   string
	client    *provider.Client
	registry  *tools.Registry
	pricing   pricing.Table
	sessions  map[string]*chatSession
	active    *chatSession
	updates   chan taggedUpdate
	loads     chan loadResult
	fold      foldMode
	sel       *selector
	mention   *mention
	slash     *slash
	file      *fileMention
	closed    closedToken
	bashDone  chan bashResult
	voice     *voiceState
	voiceDone chan voiceResult
	asyncs    chan asyncBatch
	rendered  chan renderEvent

	// asyncRunning counts the running background tasks per session.
	asyncRunning map[string]int

	// newVersion brings the result of the background release check.
	newVersion     chan string
	checkingUpdate bool
	latest         string
	dataAction     *DataAction
	kindIndex      int

	modelList     []string
	loadingModels bool
	modelsLoaded  chan modelsResult
	unread        map[string]bool
	pasting       bool
	blurred       bool
	frame         int
	glowTenths    int
	width         int
	quit          bool
}

// Run shows the TUI until the user quits. A returned DataAction must be
// carried out after the database is closed.
func Run(ctx context.Context, deps Deps) (*DataAction, error) {
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
		screen: screen, ctx: ctx, store: deps.Store, cfg: deps.Config, dir: deps.Dir, version: deps.Version,
		client: deps.Client, registry: deps.Registry, width: w, fold: foldMode(deps.Config.Fold),
		sessions: map[string]*chatSession{}, updates: make(chan taggedUpdate, 256), loads: make(chan loadResult, 4), modelsLoaded: make(chan modelsResult, 1), bashDone: make(chan bashResult, 4), voiceDone: make(chan voiceResult, 8), asyncs: make(chan asyncBatch, 4), rendered: make(chan renderEvent, 32), asyncRunning: map[string]int{}, newVersion: make(chan string, 1),
	}
	defer a.markInterruptedUnread()
	defer a.closeVoice()
	err = a.loop(ctx, deps.Pricing)
	return a.dataAction, err
}
