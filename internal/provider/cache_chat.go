package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// noCacheKey holds the endpoints that rejected prompt_cache_key.
var noCacheKey sync.Map

func rejectsCacheKey(text string) bool {
	text = strings.ToLower(text)
	return strings.Contains(text, "prompt_cache_key") || strings.Contains(text, "unknown parameter")
}

// chatRequest opens a Chat Completions stream with prompt_cache_key. When the
// endpoint rejects the field it is dropped, now and for later requests.
func (c *Client) chatRequest(ctx context.Context, model, effort string, messages []Message, toolsSchema json.RawMessage) (*http.Response, error) {
	baseURL := c.Config().BaseURL
	for {
		_, off := noCacheKey.Load(baseURL)
		payload, err := chatPayloadWithKey(model, effort, messages, toolsSchema, !off)
		if err != nil {
			return nil, err
		}
		resp, status, err := c.streamRequest(ctx, "/chat/completions", payload)
		if err == nil {
			return resp, nil
		}
		if off || status != http.StatusBadRequest || !rejectsCacheKey(err.Error()) {
			return nil, err
		}
		c.Debug("prompt_cache_key_fallback", map[string]any{"model": model, "http_status": status})
		noCacheKey.Store(baseURL, true)
	}
}
