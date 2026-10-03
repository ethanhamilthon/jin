package provider

type anthropicEvent struct {
	Type         string               `json:"type"`
	Index        int                  `json:"index"`
	Message      *anthropicStartMsg   `json:"message"`
	ContentBlock *anthropicBlockStart `json:"content_block"`
	Delta        *anthropicEventDelta `json:"delta"`
	Usage        *anthropicEventUsage `json:"usage"`
	Error        *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

type anthropicStartMsg struct {
	Role  string               `json:"role"`
	Usage *anthropicEventUsage `json:"usage"`
}

type anthropicBlockStart struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Thinking string `json:"thinking"`
	ID       string `json:"id"`
	Name     string `json:"name"`
}

type anthropicEventDelta struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	Thinking    string `json:"thinking"`
	PartialJSON string `json:"partial_json"`
}

type anthropicEventUsage struct {
	InputTokens          int  `json:"input_tokens"`
	OutputTokens         int  `json:"output_tokens"`
	CacheReadInputTokens *int `json:"cache_read_input_tokens"`
}

func applyAnthropicUsage(usage *Usage, u *anthropicEventUsage) {
	if u == nil {
		return
	}
	usage.Known = true
	if u.InputTokens > 0 {
		usage.Input = u.InputTokens
	}
	if u.OutputTokens > 0 {
		usage.Output = u.OutputTokens
	}
	if u.CacheReadInputTokens != nil {
		usage.CachedInput = *u.CacheReadInputTokens
		usage.CacheKnown = true
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
