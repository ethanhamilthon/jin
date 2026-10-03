package provider

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

func parseAnthropicStream(r io.Reader, onEvent func(StreamEvent)) (Response, error) {
	var msg Message
	var usage Usage
	calls := map[int]*ToolCall{}
	var order []int
	completed := false

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64<<10), maxProviderBody)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if eventType, ok := strings.CutPrefix(line, "event:"); ok {
			if strings.TrimSpace(eventType) == "message_stop" {
				completed = true
				break
			}
			continue
		}
		data, ok := strings.CutPrefix(line, "data:")
		if !ok {
			continue
		}
		var ev anthropicEvent
		if json.Unmarshal([]byte(strings.TrimSpace(data)), &ev) != nil {
			continue
		}
		if ev.Type == "error" {
			return Response{}, anthropicStreamError(ev)
		}
		if ev.Type == "message_stop" {
			completed = true
			break
		}
		handleAnthropicEvent(&msg, &usage, calls, &order, ev, onEvent)
	}
	if err := scanner.Err(); err != nil {
		return Response{}, transient(errors.New("stream read failed: " + err.Error()))
	}
	if !completed {
		return Response{}, transient(errors.New("chat completion stream ended before message_stop"))
	}
	for _, idx := range order {
		msg.ToolCalls = append(msg.ToolCalls, *calls[idx])
	}
	if msg.Role == "" {
		msg.Role = "assistant"
	}
	return Response{Message: msg, Usage: usage}, nil
}

// anthropicStreamError turns an error event into an error; an overloaded or
// internal provider error is worth retrying.
func anthropicStreamError(ev anthropicEvent) error {
	msg := "anthropic stream error"
	if ev.Error != nil && ev.Error.Message != "" {
		msg = ev.Error.Message
	}
	if ev.Error != nil && (ev.Error.Type == "overloaded_error" || ev.Error.Type == "api_error") {
		return transient(errors.New(msg))
	}
	return errors.New(msg)
}
