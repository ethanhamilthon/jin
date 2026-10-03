package provider

import (
	"encoding/json"
	"errors"
)

func (c *Client) debugUsage(raw []byte) {
	if c.debug == nil {
		return
	}
	var body map[string]json.RawMessage
	if json.Unmarshal(raw, &body) != nil {
		return
	}
	fields := make(map[string]any)
	for _, key := range []string{"input_tokens", "output_tokens", "total_tokens", "prompt_tokens", "completion_tokens",
		"cache_read_input_tokens", "cache_creation_input_tokens", "input_tokens_details", "output_tokens_details",
		"prompt_tokens_details", "completion_tokens_details"} {
		var value any
		if data := body[key]; data != nil && json.Unmarshal(data, &value) == nil {
			if number, ok := value.(float64); ok {
				fields[key] = number
			}
			if details, ok := value.(map[string]any); ok {
				clean := make(map[string]any)
				for _, name := range []string{"cached_tokens", "cache_write_tokens", "reasoning_tokens"} {
					if n, ok := details[name].(float64); ok {
						clean[name] = n
					}
				}
				fields[key] = clean
			}
		}
	}
	c.Debug("raw_usage", fields)
}

func (c *Client) debugResult(model string, response Response, err error) {
	if c.debug == nil {
		return
	}
	u := response.Usage
	fields := map[string]any{"model": model, "success": err == nil, "usage": u}
	if u.CacheKnown && u.Input > 0 {
		c.debug.mu.Lock()
		c.debug.input += u.Input
		c.debug.cached += u.CachedInput
		if u.CachedInput > 0 {
			c.debug.hits++
		} else {
			c.debug.misses++
		}
		fields["cache_token_ratio"] = float64(u.CachedInput) / float64(u.Input)
		fields["cache_token_ratio_total"] = float64(c.debug.cached) / float64(c.debug.input)
		fields["cache_hit_requests"], fields["cache_miss_requests"] = c.debug.hits, c.debug.misses
		c.debug.mu.Unlock()
	}
	if err != nil {
		fields["error_class"] = "provider_error"
		var status *statusError
		var tr *transientError
		if errors.As(err, &status) {
			fields["http_status"] = status.status
		}
		if errors.As(err, &tr) {
			fields["retryable"], fields["stalled"] = true, tr.stall
		}
	}
	c.Debug("attempt_end", fields)
}
