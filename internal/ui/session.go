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
	id           string
	path         string
	store        *store.DB
	provider     string
	client       *provider.Client
	persisted    bool
	agent        *core.Agent
	prompts      chan<- core.Request
	stop         context.CancelFunc
	pending      []core.Request
	history      []chatEntry
	rows         []chatRow
	input        []string
	cursor       int
	inputTop     int
	model        string
	effort       string
	title        string
	usage        store.Usage
	cache        cacheRate
	pricing      pricing.Table
	scroll       int
	width        int
	working      bool
	unread       bool
	openKind     core.UpdateKind
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
	ask        *askState
	bash       *bashState
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
