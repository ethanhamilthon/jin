package provider

import (
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) streamAnthropic(ctx context.Context, model, effort string, messages []Message, toolsSchema json.RawMessage, onEvent func(StreamEvent)) (Response, error) {
	sendThinking := effort != "" && c.modelSupportsThinking(model)
	payload, err := buildAnthropicPayload(model, effort, messages, toolsSchema, sendThinking)
	if err != nil {
		return Response{}, err
	}
	resp, status, err := c.streamRequest(ctx, "/v1/messages", payload)
	if err != nil {
		if sendThinking && status == http.StatusBadRequest {
			c.disableThinking(model)
			payload, err = buildAnthropicPayload(model, effort, messages, toolsSchema, false)
			if err != nil {
				return Response{}, err
			}
			resp, _, err = c.streamRequest(ctx, "/v1/messages", payload)
			if err != nil {
				return Response{}, err
			}
		} else {
			return Response{}, err
		}
	}
	defer resp.Body.Close()
	response, err := parseAnthropicStream(resp.Body, onEvent)
	if err != nil && ctx.Err() != nil {
		return Response{}, ctx.Err()
	}
	return response, err
}
