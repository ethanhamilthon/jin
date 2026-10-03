package provider

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"strings"
)

func (c *Client) streamOpenAI(ctx context.Context, model, effort string, messages []Message, toolsSchema json.RawMessage, onEvent func(StreamEvent)) (Response, error) {
	payload, err := chatPayload(model, effort, messages, toolsSchema)
	if err != nil {
		return Response{}, err
	}
	resp, _, err := c.streamRequest(ctx, "/chat/completions", payload)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	var message Message
	var usage Usage
	calls := map[int]*ToolCall{}
	var order []int
	completed := false

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64<<10), maxProviderBody)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		data, ok := strings.CutPrefix(line, "data:")
		if !ok {
			continue
		}
		data = strings.TrimSpace(data)
		if data == "[DONE]" {
			completed = true
			break
		}
		var chunk streamChunk
		if json.Unmarshal([]byte(data), &chunk) != nil {
			continue
		}
		if chunk.Usage != nil {
			usage = Usage{Input: chunk.Usage.PromptTokens, Output: chunk.Usage.CompletionTokens, Known: true}
			if chunk.Usage.PromptTokensDetails != nil {
				usage.CachedInput = chunk.Usage.PromptTokensDetails.CachedTokens
				usage.CacheKnown = true
			}
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		applyDelta(&message, chunk.Choices[0].Delta.Role, chunk.Choices[0].Delta.Content, chunk.Choices[0].Delta.ReasoningContent, onEvent)
		for _, tc := range chunk.Choices[0].Delta.ToolCalls {
			call, ok := calls[tc.Index]
			if !ok {
				call = &ToolCall{}
				calls[tc.Index] = call
				order = append(order, tc.Index)
			}
			mergeToolCall(call, tc)
		}
	}
	if err := scanner.Err(); err != nil {
		return Response{}, streamFailure(ctx, err)
	}
	if !completed {
		return Response{}, transient(errors.New("chat completion stream ended before [DONE]"))
	}
	for _, idx := range order {
		message.ToolCalls = append(message.ToolCalls, *calls[idx])
	}
	if message.Role == "" {
		return Response{}, errors.New("invalid chat completion stream")
	}
	return Response{Message: message, Usage: usage}, nil
}

func streamFailure(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return transient(errors.New("stream read failed: " + err.Error()))
}
