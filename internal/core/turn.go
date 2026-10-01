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
		for _, call := range answer.ToolCalls {
			if err := work.Err(); err != nil {
				return err
			}
			result := executeTool(work, ctx, call, a.registry, updates)
			toolMessage := provider.Message{Role: "tool", ToolCallID: call.ID, Content: result}
			*history = append(*history, toolMessage)
			if !sendHistory(ctx, updates, toolMessage) {
				return ctx.Err()
			}
		}
		a.compactIfNeeded(work, ctx, request, history, updates)
		for _, queued := range drainPrompts(prompts) {
			interjection := provider.Message{Role: "user", Content: queued.Prompt}
			*history = append(*history, interjection)
			if !sendHistory(ctx, updates, interjection) {
				return ctx.Err()
			}
			request.Model, request.Effort, request.Window = queued.Model, queued.Effort, queued.Window
		}
	}
}
