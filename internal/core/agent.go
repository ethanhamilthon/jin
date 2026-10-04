package core

import (
	"context"
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
	size, mark   int
	answers      chan []string

	compactText, handoffText string
	background               func(tools.Adoption) (string, error)
	gate                     func(provider.Usage) error
	gateUsage                provider.Usage
	waiting                  atomic.Int32
	detachMu                 sync.Mutex
	detach                   chan struct{}

	// Used only by the Run goroutine after it starts.
	refresh              Refresher
	clock                func() time.Time
	renderedAt, lastDone time.Time
	refreshDue           bool
	refreshNote          string
}

func NewAgent(client *provider.Client, systemPrompt string, registry *tools.Registry) *Agent {
	return &Agent{client: client, systemPrompt: systemPrompt, registry: registry, answers: make(chan []string, 1), renderedAt: time.Now()}
}

// SetSystemPrompt replaces the system prompt, which is the first message of
// the history, and notes the time it was rendered. Call it before Run.
func (a *Agent) SetSystemPrompt(text string) { a.systemPrompt, a.renderedAt = text, a.now() }

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

func (a *Agent) Interrupt() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancelTurn != nil {
		a.cancelTurn()
	}
}

// Answer delivers the user's answers to a pending ask_user call.
func (a *Agent) Answer(answers []string) {
	select {
	case a.answers <- answers:
	default:
	}
}
