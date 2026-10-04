package core

import (
	"context"
	"errors"

	"jin/internal/provider"
)

// SetRequestGate sets a check before each answer or side request. It receives
// the previous response's usage (zero before the first request); an error
// ends the turn without sending the request. Call it before Run.
func (a *Agent) SetRequestGate(gate func(last provider.Usage) error) { a.gate = gate }

type requestGateError struct{ error }

func (e *requestGateError) Unwrap() error { return e.error }

func isGateError(err error) bool {
	var refused *requestGateError
	return errors.As(err, &refused)
}

func (a *Agent) gatedStream(ctx context.Context, request Request, history []provider.Message, onEvent func(provider.StreamEvent)) (provider.Response, error) {
	if a.gate != nil {
		if err := a.gate(a.gateUsage); err != nil {
			return provider.Response{}, &requestGateError{err}
		}
	}
	a.gateUsage = provider.Usage{}
	response, err := a.client.Stream(ctx, request.Model, request.Effort, history, a.registry.SchemaJSON(), onEvent)
	a.gateUsage = response.Usage
	return response, err
}
