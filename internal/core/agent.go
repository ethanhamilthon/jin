package core

import (
	"context"
	"errors"
	"strings"
	"sync"

	"jin/internal/provider"
	"jin/internal/tools"
)

type Request struct {
	Prompt, Model, Effort string
}

type Agent struct {
	client       *provider.Client
	systemPrompt string
	registry     *tools.Registry
	mu           sync.Mutex
	cancelTurn   context.CancelFunc
}

func NewAgent(client *provider.Client, systemPrompt string, registry *tools.Registry) *Agent {
	return &Agent{client: client, systemPrompt: systemPrompt, registry: registry}
}

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
			if strings.TrimSpace(request.Prompt) == "" {
				continue
			}
			if !a.turn(ctx, request, &history, prompts, updates) {
				return
			}
		}
	}
}

func (a *Agent) turn(ctx context.Context, request Request, history *[]provider.Message, prompts <-chan Request, updates chan<- Update) bool {
	userMessage := provider.Message{Role: "user", Content: request.Prompt}
	*history = append(*history, userMessage)
	if !sendHistory(ctx, updates, userMessage) || !sendUpdate(ctx, updates, UpdateWorking, "") {
		return false
	}
	turnCtx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.cancelTurn = cancel
	a.mu.Unlock()
	err := a.answer(turnCtx, ctx, request, history, prompts, updates)
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
	return sendUpdate(ctx, updates, UpdateDone, "")
}

var errNoModel = errors.New("no model selected: press m in NORMAL mode")
