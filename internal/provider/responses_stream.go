package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

func (c *Client) streamResponses(ctx context.Context, model, effort string, messages []Message, toolsSchema json.RawMessage, onEvent func(StreamEvent)) (Response, error) {
	payload, err := buildResponsesPayload(model, effort, messages, toolsSchema)
	if err != nil {
		return Response{}, err
	}
	resp, _, err := c.streamRequest(ctx, "/responses", payload)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	var result Response
	var items []json.RawMessage
	completed := false
	err = readSSE(resp.Body, func(data []byte) (bool, error) {
		var ev responsesEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return false, errors.New("invalid Responses stream event")
		}
		switch ev.Type {
		case "response.output_text.delta", "response.refusal.delta":
			onEvent(StreamEvent{Kind: DeltaContent, Text: ev.Delta})
		case "response.reasoning_summary_text.delta":
			onEvent(StreamEvent{Kind: DeltaReasoning, Text: ev.Delta})
		case "response.output_item.done":
			items = append(items, ev.Item)
		case "response.completed":
			if ev.Response == nil {
				return false, errors.New("Responses completion is missing response")
			}
			if ev.Response.Output != nil {
				items = ev.Response.Output
			}
			result.Usage = ev.Response.Usage.usage()
			var envelope struct {
				Response struct {
					Usage json.RawMessage `json:"usage"`
				} `json:"response"`
			}
			_ = json.Unmarshal(data, &envelope)
			c.debugUsage(envelope.Response.Usage)
			var err error
			result.Message, err = responsesOutput(items, model)
			completed = err == nil
			return true, err
		case "response.failed", "response.incomplete", "error":
			if ev.Response != nil {
				result.Usage = ev.Response.Usage.usage()
			}
			return true, responsesStreamError(ev)
		}
		return false, nil
	})
	if err == nil && !completed {
		err = transient(errors.New("Responses stream ended before response.completed"))
	}
	return result, err
}

func responsesStreamError(ev responsesEvent) error {
	code, message := ev.Code, ev.Message
	if ev.Response != nil {
		if ev.Response.Error != nil {
			code, message = ev.Response.Error.Code, ev.Response.Error.Message
		}
		if ev.Response.IncompleteDetails != nil {
			code = ev.Response.IncompleteDetails.Reason
		}
	}
	err := fmt.Errorf("Responses %s: %s", ev.Type, code+" "+message)
	if code == "server_error" || code == "rate_limit_exceeded" {
		return transient(err)
	}
	return err
}
