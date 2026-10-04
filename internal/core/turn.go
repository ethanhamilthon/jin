package core

import (
	"context"
	"errors"

	"jin/internal/provider"
)

// answer drives one turn until the model replies without tool calls. work is
// cancelled on interrupt; ctx outlives it so history still reaches the UI.
// Prompts queued while tools run are folded in before the next model call.
// A request refused as too long is retried once after making room.
func (a *Agent) answer(work, ctx context.Context, request Request, history *[]provider.Message, prompts <-chan Request, updates chan<- Update) error {
	recovered := false
	for {
		if request.Model == "" {
			return errNoModel
		}
		response, err := a.client.Stream(work, request.Model, request.Effort, *history, a.registry.SchemaJSON(), func(event provider.StreamEvent) {
			switch event.Kind {
			case provider.Notice:
				sendUpdate(work, updates, UpdateInfo, event.Text)
			case provider.DeltaReasoning:
				sendDelta(work, updates, UpdateReasoningDelta, event.Text)
			default:
				sendDelta(work, updates, UpdateAssistantDelta, event.Text)
			}
		})
		if err != nil && !recovered && provider.IsContextOverflow(err) && work.Err() == nil {
			recovered = true
			if err := a.recoverOverflow(work, ctx, request, history, updates); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		recovered = false
		answer := response.Message
		sendUsage(ctx, updates, request.Model, response.Usage)
		if len(answer.ToolCalls) == 0 && answer.Content == "" {
			return errors.New("assistant returned no response")
		}
		*history = append(*history, answer)
		a.reported(response.Usage, *history)
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
		a.refreshSystem(work, *history)
		for _, queued := range drainPrompts(prompts) {
			a.consumedRequest(queued)
			interjection := provider.Message{Role: "user", Content: a.takeRefreshNote() + queued.Prompt}
			*history = append(*history, interjection)
			if !sendHistory(ctx, updates, interjection) {
				return ctx.Err()
			}
			request.Model, request.Effort, request.Window, request.NoVision = queued.Model, queued.Effort, queued.Window, queued.NoVision
		}
	}
}
