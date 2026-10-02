package core

import (
	"context"
	"errors"

	"jin/internal/provider"
)

// answer drives one turn until the model replies without tool calls. work is
// cancelled on interrupt; ctx outlives it so history still reaches the UI.
// Prompts queued while tools run are folded in before the next model call.
func (a *Agent) answer(work, ctx context.Context, request Request, history *[]provider.Message, prompts <-chan Request, updates chan<- Update) error {
	for {
		if request.Model == "" {
			return errNoModel
		}
		response, err := a.client.Stream(work, request.Model, request.Effort, *history, a.registry.SchemaJSON(), func(event provider.StreamEvent) {
			kind := UpdateAssistantDelta
			if event.Kind == provider.DeltaReasoning {
				kind = UpdateReasoningDelta
			}
			sendDelta(work, updates, kind, event.Text)
		})
		if err != nil {
			return err
		}
		answer := response.Message
		sendUsage(ctx, updates, request.Model, response.Usage)
		if response.Usage.Known {
			a.size = response.Usage.Input + response.Usage.Output
		}
		if len(answer.ToolCalls) == 0 && answer.Content == "" {
			return errors.New("assistant returned no response")
		}
		*history = append(*history, answer)
		if !sendHistory(ctx, updates, answer) {
			return ctx.Err()
		}
		if len(answer.ToolCalls) == 0 {
			return nil
		}
		if err := a.runTools(work, ctx, request, answer.ToolCalls, history, updates); err != nil {
			return err
		}
		a.compactIfNeeded(work, ctx, request, history, updates)
		for _, queued := range drainPrompts(prompts) {
			a.consumedRequest(queued)
			interjection := provider.Message{Role: "user", Content: queued.Prompt}
			*history = append(*history, interjection)
			if !sendHistory(ctx, updates, interjection) {
				return ctx.Err()
			}
			request.Model, request.Effort, request.Window, request.NoVision = queued.Model, queued.Effort, queued.Window, queued.NoVision
		}
	}
}
