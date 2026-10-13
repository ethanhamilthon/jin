package provider

import (
	"context"
	"encoding/json"
	"errors"
)

func (c *Client) autoStream(ctx context.Context, model, effort string, messages []Message, tools json.RawMessage, onEvent func(StreamEvent), alternateUsed *bool, active *string) (Response, error) {
	if *active == "" {
		*active = c.protocolFor(model, c.Config())
	}
	protocol := *active
	response, err := streamProtocol(ctx, c, protocol, model, effort, messages, tools, onEvent)
	if err == nil {
		c.rememberProtocol(model, protocol)
		return response, nil
	}
	var status *statusError
	if *alternateUsed || !errors.As(err, &status) || !protocolRejected(status) || ctx.Err() != nil {
		return response, err
	}
	*alternateUsed = true
	alternate := KindOpenAI
	if protocol == KindOpenAI {
		alternate = KindResponses
	}
	*active = alternate
	onEvent(StreamEvent{Kind: Notice, Text: "Protocol rejected. Retrying with " + alternate + "."})
	response, alternateErr := streamProtocol(ctx, c, alternate, model, effort, messages, tools, onEvent)
	if alternateErr == nil {
		c.rememberProtocol(model, alternate)
		return response, nil
	}
	return response, errors.New("both OpenAI protocols failed: " + err.Error() + "; " + alternateErr.Error())
}

func streamProtocol(ctx context.Context, c *Client, protocol, model, effort string, messages []Message, tools json.RawMessage, onEvent func(StreamEvent)) (Response, error) {
	if protocol == KindResponses {
		return c.streamResponses(ctx, model, effort, messages, tools, onEvent)
	}
	return c.streamOpenAI(ctx, model, effort, messages, tools, onEvent)
}
