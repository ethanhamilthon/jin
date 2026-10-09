package ui

import (
	"context"

	"jin/internal/core"
	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
	"jin/internal/todo"
)

// chatSession is one conversation with its own backend loop. It keeps running
// in the background while another session is focused.
type chatSession struct {
	id       string
	path     string
	store    *store.DB
	provider string
	client   *provider.Client
	// providerMissing is true when the saved provider was deleted: the
	// session cannot send until the user picks another one.
	providerMissing bool
	// readOnlyPID is the process that uses the session; 0 when it is ours.
	readOnlyPID int
	// models is the model list of the provider named by modelsFor.
	models        []string
	modelsFor     string
	persisted     bool
	agent         *core.Agent
	toolNames     []string
	prompts       chan<- core.Request
	stop          context.CancelFunc
	backendDone   <-chan struct{}
	pending       []core.Request
	history       []chatEntry
	rows          []chatRow
	input         []string
	draftRevision uint64
	cursor        int
	inputTop      int
	model         string
	effort        string
	title         string
	usage         store.Usage
	cache         cacheRate
	pricing       pricing.Table
	scroll        int
	width         int
	working       bool
	inflight      int
	unread        bool
	openKind      core.UpdateKind
	// continueAnswer joins the next reply to the answer above: it follows the background tasks note.
	continueAnswer bool
	// suggestion is the user's likely next request from tell_user, shown in the empty input.
	suggestion   string
	openRowStart int
	stream       streamState
	view         viewport
	selection    textSelection
	fold         foldMode
	todos        []todo.Item
	// changeTurn numbers the file changes of the running turn for /undo;
	// undoNote tells the model about an undo with the next message.
	changeTurn int
	undoNote   string
	todoTop    int
	// todoFolded hides the pinned list; todoRule is the row of its rule plus one.
	todoFolded  bool
	todoRule    int
	todoPressed bool
	ask         *askState
	bash        *bashState
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
