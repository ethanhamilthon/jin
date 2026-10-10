package ui

import (
	"context"
	"jin/internal/daemon"
	"jin/internal/session"

	"jin/internal/core"
	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/sources"
	"jin/internal/store"
)

// chatSession is one conversation with its own backend loop. It keeps running
// in the background while another session is focused.
type chatSession struct {
	backend *daemon.Client
	// remoteState is the daemon's view of the session; queued and paused
	// mirror its queue so the status line can show it.
	remoteState   session.State
	remoteEntries []session.Entry
	// remoteMap is the chat history index of each daemon entry: the client
	// appends its own rows too, so the two lists drift apart.
	remoteMap []int
	remoteSeq int64
	id        string
	path      string
	store     *store.DB
	provider  string
	client    *provider.Client
	// providerMissing is true when the saved provider was deleted: the
	// session cannot send until the user picks another one.
	providerMissing bool
	// readOnlyPID is the process that uses the session; 0 when it is ours.
	readOnlyPID int
	// models is the model list of the provider named by modelsFor.
	models         []string
	modelsFor      string
	modelChoices   []sources.Model
	modelsRevision string
	persisted      bool
	agent          *core.Agent
	toolNames      []string
	prompts        chan<- core.Request
	stop           context.CancelFunc
	backendDone    <-chan struct{}
	pending        []core.Request
	history        []chatEntry
	rows           []chatRow
	input          []string
	draftRevision  uint64
	cursor         int
	inputTop       int
	model          string
	effort         string
	title          string
	titledAt       int
	usage          store.Usage
	cache          cacheRate
	pricing        pricing.Table
	scroll         int
	width          int
	working        bool
	inflight       int
	unread         bool
	openKind       core.UpdateKind
	// continueAnswer joins the next reply to the answer above: it follows the background tasks note.
	continueAnswer bool
	openRowStart   int
	stream         streamState
	view           viewport
	selection      textSelection
	fold           foldMode
	ask            *askState
	bash           *bashState
	// ready is false while the session starts: its commands run in the
	// background, its agent is not running and its input is closed.
	ready  bool
	render *rendering
	// promptBodies are the #prompts, with their commands run, as of the start.
	promptBodies map[string]string
	// customSystem is true when ~/.jin/system-prompt.md was used.
	customSystem bool
	// initial, runCtx and updatesOut are what the agent needs to start.
	initial    []provider.Message
	runCtx     context.Context
	updatesOut chan core.Update
	requests   <-chan core.Request
}

// viewport is where the timeline was last drawn, for mouse hit-testing.
type viewport struct {
	first, height int
}
