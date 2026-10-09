package session

import (
	"context"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/store"
	"jin/internal/tools"
)

// Session is one live conversation with its own agent loop. The Manager's
// lock guards every field.
type Session struct {
	m        *Manager
	id, path string
	provider string
	client   *provider.Client
	agent    *core.Agent
	names    []string
	prompts  chan core.Request
	requests chan core.Request
	updates  chan core.Update
	runCtx   context.Context
	stop     context.CancelFunc
	done     chan struct{}
	initial  []provider.Message

	pending  []core.Request
	inflight int
	entries  []Entry
	open     core.UpdateKind
	attempt  []int
	joinNext bool

	model, effort, title string
	usage                store.Usage
	cache                *int
	persisted, ready     bool
	working, unread      bool
	readOnlyPID          int
	providerMissing      bool
	render               *rendering
	bodies               map[string]string
	ask                  []tools.Question
	changeTurn           int
	undoNote             string
	shell                context.CancelFunc
	draft                string
	draftRev             int
	intro                *Intro
}

// rendering is the start-up or reload of the session's prompts, whose
// commands run in the background.
type rendering struct {
	cancel  context.CancelFunc
	reload  bool
	loading []string
}

func (s *Session) ID() string { return s.id }

func (s *Session) busy() bool {
	return s.working || s.inflight > 0 || len(s.pending) > 0 || s.render != nil || s.shell != nil
}
