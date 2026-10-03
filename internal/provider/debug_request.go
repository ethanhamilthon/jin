package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var debugSequence atomic.Uint64

type debugRequestKey struct{}
type debugRequestInfo struct {
	id      string
	started time.Time
}

func processID() string { return strconv.Itoa(os.Getpid()) }
func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (c *Client) debugRequest(ctx context.Context, path string, payload []byte) context.Context {
	if c.debug == nil {
		return ctx
	}
	id := processID() + "-" + strconv.FormatUint(debugSequence.Add(1), 10)
	var body map[string]json.RawMessage
	_ = json.Unmarshal(payload, &body)
	var model string
	_ = json.Unmarshal(body["model"], &model)
	history := body["messages"]
	if history == nil {
		history = body["input"]
	}
	var items []json.RawMessage
	_ = json.Unmarshal(history, &items)
	hashes := []string{digest(body["tools"]), digest(body["system"]), digest(body["reasoning"]), digest(body["thinking"]), digest(body["reasoning_effort"]), digest(body["output_config"])}
	for _, item := range items {
		hashes = append(hashes, digest(item))
	}
	key := c.Config().BaseURL + path + model
	c.debug.mu.Lock()
	previous, common := c.debug.previous[key], 0
	for common < len(previous) && common < len(hashes) && previous[common] == hashes[common] {
		common++
	}
	c.debug.previous[key] = hashes
	c.debug.mu.Unlock()
	c.Debug("request", map[string]any{"request_id": id, "path": path, "model": model,
		"payload_bytes": len(payload), "history_items": len(items), "shared_prefix_items": max(0, common-6),
		"prefix_unchanged": len(previous) > 0 && common == len(previous),
		"tools_hash":       hashes[0], "system_hash": hashes[1], "history_hash": digest(history)})
	return context.WithValue(ctx, debugRequestKey{}, debugRequestInfo{id: id, started: time.Now()})
}

func (c *Client) debugHTTP(ctx context.Context, event string, fields map[string]any) {
	if info, ok := ctx.Value(debugRequestKey{}).(debugRequestInfo); ok {
		fields["request_id"], fields["elapsed_ms"] = info.id, time.Since(info.started).Milliseconds()
		c.Debug(event, fields)
	}
}
