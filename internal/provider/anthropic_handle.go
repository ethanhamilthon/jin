package provider

func applyAnthropicUsage(usage *Usage, u *anthropicEventUsage) {
	if u == nil {
		return
	}
	usage.Known = true
	uncached := usage.Input - usage.CachedInput - usage.CacheWriteInput
	if u.InputTokens != nil {
		uncached = *u.InputTokens
	}
	if u.OutputTokens != nil {
		usage.Output = *u.OutputTokens
	}
	if u.CacheReadInputTokens != nil {
		usage.CachedInput, usage.CacheKnown = *u.CacheReadInputTokens, true
	}
	if u.CacheCreationInputTokens != nil {
		usage.CacheWriteInput, usage.CacheWriteKnown = *u.CacheCreationInputTokens, true
	}
	usage.Input = uncached + usage.CachedInput + usage.CacheWriteInput
	if u.OutputTokensDetails != nil && u.OutputTokensDetails.Reasoning != nil {
		usage.Reasoning, usage.ReasoningKnown = *u.OutputTokensDetails.Reasoning, true
	}
}

func handleAnthropicEvent(msg *Message, usage *Usage, calls map[int]*ToolCall, order *[]int, ev anthropicEvent, onEvent func(StreamEvent)) {
	switch ev.Type {
	case "message_start":
		if ev.Message != nil {
			if ev.Message.Role != "" {
				msg.Role = ev.Message.Role
			}
			applyAnthropicUsage(usage, ev.Message.Usage)
		}
	case "content_block_start":
		if ev.ContentBlock != nil && ev.ContentBlock.Type == "tool_use" {
			c := getAnthropicCall(calls, order, ev.Index)
			c.ID, c.Function.Name = ev.ContentBlock.ID, ev.ContentBlock.Name
			if input := string(ev.ContentBlock.Input); input != "" && input != "{}" {
				c.Function.Arguments = input
			}
		} else if ev.ContentBlock != nil {
			applyDelta(msg, "", ev.ContentBlock.Text, ev.ContentBlock.Thinking, onEvent)
		}
	case "content_block_delta":
		if ev.Delta != nil {
			handleAnthropicDelta(msg, calls, order, ev.Index, ev.Delta, onEvent)
		}
	case "message_delta":
		applyAnthropicUsage(usage, ev.Usage)
	}
}

func handleAnthropicDelta(msg *Message, calls map[int]*ToolCall, order *[]int, idx int, d *anthropicEventDelta, onEvent func(StreamEvent)) {
	switch d.Type {
	case "text_delta":
		if d.Text != "" {
			msg.Content += d.Text
			onEvent(StreamEvent{Kind: DeltaContent, Text: d.Text})
		}
	case "thinking_delta":
		if d.Thinking != "" {
			msg.ReasoningContent += d.Thinking
			onEvent(StreamEvent{Kind: DeltaReasoning, Text: d.Thinking})
		}
	case "input_json_delta":
		getAnthropicCall(calls, order, idx).Function.Arguments += d.PartialJSON
	}
}

func getAnthropicCall(calls map[int]*ToolCall, order *[]int, idx int) *ToolCall {
	call, ok := calls[idx]
	if !ok {
		call = &ToolCall{Type: "function"}
		calls[idx] = call
		*order = append(*order, idx)
	}
	return call
}
