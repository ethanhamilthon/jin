package provider

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Role             string           `json:"role"`
			Content          string           `json:"content"`
			ReasoningContent string           `json:"reasoning_content"`
			ToolCalls        []streamToolCall `json:"tool_calls"`
		} `json:"delta"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens        int `json:"prompt_tokens"`
		CompletionTokens    int `json:"completion_tokens"`
		PromptTokensDetails *struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
	} `json:"usage"`
}

type streamToolCall struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

func applyDelta(message *Message, role, content, reasoning string, onEvent func(StreamEvent)) {
	if role != "" {
		message.Role = role
	}
	if content != "" {
		message.Content += content
		onEvent(StreamEvent{Kind: DeltaContent, Text: content})
	}
	if reasoning != "" {
		message.ReasoningContent += reasoning
		onEvent(StreamEvent{Kind: DeltaReasoning, Text: reasoning})
	}
}

func mergeToolCall(call *ToolCall, delta streamToolCall) {
	if delta.ID != "" {
		call.ID = delta.ID
	}
	if delta.Type != "" {
		call.Type = delta.Type
	}
	if delta.Function.Name != "" {
		call.Function.Name = delta.Function.Name
	}
	call.Function.Arguments += delta.Function.Arguments
}
