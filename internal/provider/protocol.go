package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type autoState struct {
	active        string
	alternateUsed bool
	streamed      bool
}

func (c *Client) autoStream(ctx context.Context, model, effort string, messages []Message, tools json.RawMessage, onEvent func(StreamEvent), st *autoState) (Response, error) {
	cfg := c.Config()
	if st.active == "" {
		st.active = c.protocolFor(model, cfg)
	}
	emit := func(e StreamEvent) {
		if e.Kind == DeltaContent || e.Kind == DeltaReasoning {
			st.streamed = true
		}
		onEvent(e)
	}
	protocol := st.active
	response, err := streamProtocol(ctx, c, protocol, model, effort, messages, tools, emit)
	if err == nil {
		c.rememberProtocol(cfg, model, protocol)
		return response, nil
	}
	var status *statusError
	if st.alternateUsed || st.streamed || !errors.As(err, &status) || !protocolRejected(status) || ctx.Err() != nil {
		return response, err
	}
	st.alternateUsed = true
	alternate := KindOpenAI
	if protocol == KindOpenAI {
		alternate = KindResponses
	}
	st.active = alternate
	onEvent(StreamEvent{Kind: Notice, Text: "Protocol rejected. Retrying with " + alternate + "."})
	response, alternateErr := streamProtocol(ctx, c, alternate, model, effort, messages, tools, emit)
	if alternateErr == nil {
		c.rememberProtocol(cfg, model, alternate)
		return response, nil
	}
	return response, fmt.Errorf("both OpenAI protocols failed (%s): %w", err.Error(), alternateErr)
}

func streamProtocol(ctx context.Context, c *Client, protocol, model, effort string, messages []Message, tools json.RawMessage, onEvent func(StreamEvent)) (Response, error) {
	if protocol == KindResponses {
		return c.streamResponses(ctx, model, effort, messages, tools, onEvent)
	}
	return c.streamOpenAI(ctx, model, effort, messages, tools, onEvent)
}
