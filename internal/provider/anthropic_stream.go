package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func (c *Client) streamAnthropic(ctx context.Context, model, effort string, messages []Message, toolsSchema json.RawMessage, onEvent func(StreamEvent)) (Response, error) {
	key := c.featureKey(model)
	known := c.knownFeatures(key)
	opts := anthropicOptions{thinking: effort != "" && !known.noThinking, cache: !known.noCache}
	var resp *http.Response
	for {
		payload, err := buildAnthropicPayload(model, effort, messages, toolsSchema, opts)
		if err != nil {
			return Response{}, err
		}
		var status int
		resp, status, err = c.streamRequest(ctx, "/v1/messages", payload)
		if err == nil {
			break
		}
		if status != http.StatusBadRequest {
			return Response{}, err
		}
		// Some proxies set their own cache_control; a conflicting top-level one is rejected.
		if opts.cache && strings.Contains(err.Error(), "cache_control") {
			c.Debug("cache_control_fallback", map[string]any{"model": model, "http_status": status})
			c.updateFeatures(key, func(f *anthropicFeatures) { f.noCache = true })
			opts.cache = false
			continue
		}
		if !opts.thinking || !mentionsThinking(err.Error()) {
			return Response{}, err
		}
		c.Debug("thinking_fallback", map[string]any{"model": model, "http_status": status})
		c.updateFeatures(key, func(f *anthropicFeatures) { f.noThinking = true })
		opts.thinking = false
	}
	defer resp.Body.Close()
	response, err := parseAnthropicStream(resp.Body, onEvent, c.debugUsage)
	if err != nil && ctx.Err() != nil {
		return Response{}, ctx.Err()
	}
	if response.Message.Native != nil {
		response.Message.Native.Model = model
	}
	return response, err
}
