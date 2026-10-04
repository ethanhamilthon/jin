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
	asyncPaths         chan []string
	updates            chan taggedUpdate
	loads              chan loadResult
	fold               foldMode
	sel                *selector
	mention            *mention
	slash              *slash
	file               *fileMention
	closed             closedToken
	bashDone           chan bashResult
	asyncs             chan asyncBatch
	rendered           chan renderEvent

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
