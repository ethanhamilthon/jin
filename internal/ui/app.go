package ui

import (
	"context"
	"jin/internal/daemon"
	"jin/internal/session"
	"jin/internal/tasks"

	"github.com/gdamore/tcell/v3"

	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
	"jin/internal/tools"
)

type Deps struct {
	Backend  *daemon.Client
	Store    *store.DB
	Config   store.Config
	Client   *provider.Client
	Registry *tools.Registry
	Pricing  <-chan pricing.Table
	Dir      string
	Version  string
}

type app struct {
	backend            *daemon.Client
	backendEvents      <-chan session.Event
	screen             tcell.Screen
	ctx                context.Context
	cancel             context.CancelFunc
	store              *store.DB
	cfg                store.Config
	dir                string
	version            string
	client             *provider.Client
	registry           *tools.Registry
	pricing            pricing.Table
	sessions           map[string]*chatSession
	active             *chatSession
	panes              *paneNode
	focused            *paneNode
	lastProjectSession map[string]string
	updates            chan taggedUpdate
	loads              chan loadResult
	fold               foldMode
	sel                *selector
	mention            *mention
	slash              *slash
	file               *fileMention
	closed             closedToken
	bashDone           chan bashResult
	taskEvents         <-chan tasks.Event
	heldTasks          []tasks.Event
	rendered           chan renderEvent

	// tasksRunning counts the running background tasks per session.
	tasksRunning map[string]int

	// newVersion brings the result of the background release check.
	newVersion     chan string
	checkingUpdate bool
	latest         string
	dataAction     *DataAction
	kindIndex      int
	// onboardMode is the first-run choice: "" before it, then "api" or "subscription".
	onboardMode string

	modelList     []string
	loadingModels bool
	modelsLoaded  chan modelsResult
	// titles brings the session names made in the background.
	titles     chan titleResult
	unread     map[string]bool
	pasting    bool
	blurred    bool
	frame      int
	glowTenths int
	width      int
	quit       bool
}
