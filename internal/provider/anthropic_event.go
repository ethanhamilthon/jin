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
