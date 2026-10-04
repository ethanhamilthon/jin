package provider

import (
	"context"
	"encoding/json"
)

type StreamEventKind int

const (
	DeltaContent StreamEventKind = iota
	DeltaReasoning
	// Notice is a status line, such as a retry after a provider error.
	Notice
	// Reset says the attempt that streamed so far failed and is retried: what
	// it sent is void. Usage is what the failed attempt reported, if anything.
	Reset
)

type StreamEvent struct {
	Kind  StreamEventKind
	Text  string
	Usage Usage
}

// Stream reports content and reasoning fragments as they arrive and returns
// the assembled message once the stream ends. Tool call arguments are
// assembled internally: partial JSON is not worth showing.
// Transient failures are retried; each retry sends a Reset, then a Notice.
func (c *Client) Stream(ctx context.Context, model, effort string, messages []Message, toolsSchema json.RawMessage, onEvent func(StreamEvent)) (Response, error) {
	return c.withRetry(ctx, effort, onEvent, func(ctx context.Context) (response Response, err error) {
		defer func() { c.debugResult(model, response, err) }()
		switch c.Config().Kind {
		case KindAnthropic:
			return c.streamAnthropic(ctx, model, effort, messages, toolsSchema, onEvent)
		case KindResponses:
			return c.streamResponses(ctx, model, effort, messages, toolsSchema, onEvent)
		default:
			return c.streamOpenAI(ctx, model, effort, messages, toolsSchema, onEvent)
		}
	})
}
