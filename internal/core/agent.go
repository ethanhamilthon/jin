package core

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"jin/internal/provider"
	"jin/internal/sysprompt"
	"jin/internal/tools"
)

type Agent struct {
	client       *provider.Client
	systemPrompt string
	registry     *tools.Registry
	mu           sync.Mutex
	cancelTurn   context.CancelFunc
	size         int
	answers      chan []string

	compactText, handoffText string
	background               func(tools.Adoption) (string, error)
	waiting                  atomic.Int32
	detachMu                 sync.Mutex
	detach                   chan struct{}
}

func NewAgent(client *provider.Client, systemPrompt string, registry *tools.Registry) *Agent {
	return &Agent{client: client, systemPrompt: systemPrompt, registry: registry, answers: make(chan []string, 1)}
}

// SetSystemPrompt replaces the system prompt, which is the first message of
// the history. Call it before Run.
func (a *Agent) SetSystemPrompt(text string) { a.systemPrompt = text }

// SetSidePrompts sets the instructions for compaction and handoff. An empty
// text means the built-in default.
func (a *Agent) SetSidePrompts(compact, handoff string) {
	a.compactText, a.handoffText = compact, handoff
}

func (a *Agent) compactPrompt() string {
	if strings.TrimSpace(a.compactText) != "" {
		return a.compactText
	}
	return sysprompt.Defaults().Compact
}

func (a *Agent) handoffPrompt() string {
	if strings.TrimSpace(a.handoffText) != "" {
		return a.handoffText
	}
	return sysprompt.Defaults().Handoff
}

// SetBackground tells the agent how to hand a bash command over to the async
// daemon. Without it a command whose time is up is killed.
func (a *Agent) SetBackground(adopt func(tools.Adoption) (string, error)) {
	a.background = adopt
}

// Expect counts a user message that is on its way to the agent. Call it
// before the message is queued; a tool that starts or is running meanwhile
// moves to the background instead of making the user wait.
func (a *Agent) Expect() {
	if a != nil {
		a.waiting.Add(1)
	}
}

func (a *Agent) consumed(n int32) {
	if a.waiting.Add(-n) < 0 {
		a.waiting.Store(0)
	}
}

// DetachTools asks the running tool calls to move to the background now. It
// is called after a user message was queued, so a long bash command does not
// hold the message back.
func (a *Agent) DetachTools() {
	if a == nil {
		return
	}
	a.detachMu.Lock()
	defer a.detachMu.Unlock()
	if a.detach != nil {
		close(a.detach)
		a.detach = nil
	}
}

// toolDetach is the channel a tool call watches: it closes when a user
// message is waiting. A message that was already waiting when the tool
// starts closes it after a short moment, so quick tools still finish.
func (a *Agent) toolDetach() <-chan struct{} {
	a.detachMu.Lock()
	defer a.detachMu.Unlock()
	if a.detach == nil {
		a.detach = make(chan struct{})
	}
	ch := a.detach
	if a.waiting.Load() > 0 {
		go func() {
			time.Sleep(detachDelay)
			a.DetachTools()
		}()
	}
	return ch
}

// detachDelay is how long a tool may run while a user message already waits.
const detachDelay = time.Second

func (a *Agent) Interrupt() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancelTurn != nil {
		a.cancelTurn()
	}
}

func (a *Agent) Run(ctx context.Context, initial []provider.Message, prompts <-chan Request, updates chan<- Update) {
	defer close(updates)
	history := []provider.Message{{Role: "system", Content: a.systemPrompt}}
	history = append(history, initial...)
	for {
		select {
		case <-ctx.Done():
			return
		case request, ok := <-prompts:
			if !ok {
				return
			}
			a.consumedRequest(request)
			if request.blank() {
				continue
			}
			if !a.turn(ctx, request, &history, prompts, updates) {
				return
			}
		}
	}
}

func (a *Agent) turn(ctx context.Context, request Request, history *[]provider.Message, prompts <-chan Request, updates chan<- Update) bool {
	if !sendUpdate(ctx, updates, UpdateWorking, "") {
		return false
	}
	turnCtx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.cancelTurn = cancel
	a.mu.Unlock()
	err := a.perform(turnCtx, ctx, request, history, prompts, updates)
	a.mu.Lock()
	a.cancelTurn = nil
	a.mu.Unlock()
	interrupted := turnCtx.Err() != nil
	cancel()
	if ctx.Err() != nil {
		return false
	}
	switch {
	case interrupted:
		for _, msg := range InterruptedToolMessages(*history) {
			*history = append(*history, msg)
			sendHistory(ctx, updates, msg)
		}
		sendUpdate(ctx, updates, UpdateInfo, "Interrupted")
	case err != nil:
		sendUpdate(ctx, updates, UpdateError, err.Error())
	}
	return sendDone(ctx, updates, err == nil && !interrupted && request.Kind == RequestPrompt)
}

func (a *Agent) perform(work, ctx context.Context, request Request, history *[]provider.Message, prompts <-chan Request, updates chan<- Update) error {
	switch request.Kind {
	case RequestCompact:
		return a.compact(work, ctx, request, history, updates)
	case RequestHandoff:
		return a.handoff(work, ctx, request, history, updates)
	}
	a.compactIfNeeded(work, ctx, request, history, updates)
	userMessage := provider.Message{Role: "user", Content: request.Prompt}
	*history = append(*history, userMessage)
	if !sendHistory(ctx, updates, userMessage) {
		return ctx.Err()
	}
	return a.answer(work, ctx, request, history, prompts, updates)
}

var errNoModel = errors.New("no model selected: press Esc, open Settings, then Select model")

// Answer delivers the user's answers to a pending ask_user call.
func (a *Agent) Answer(answers []string) {
	select {
	case a.answers <- answers:
	default:
	}
}
