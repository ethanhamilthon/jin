package provider

import (
	"encoding/json"
	"errors"
)

type anthropicPayload struct {
	Model        string           `json:"model"`
	MaxTokens    int              `json:"max_tokens"`
	System       string           `json:"system,omitempty"`
	Messages     []anthropicMsg   `json:"messages"`
	Tools        []anthropicTool  `json:"tools,omitempty"`
	Stream       bool             `json:"stream"`
	Thinking     *anthropicThink  `json:"thinking,omitempty"`
	OutputConfig *anthropicOutput `json:"output_config,omitempty"`
}

type anthropicTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type anthropicToolUseBlock struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type anthropicToolResultBlock struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

type anthropicThink struct {
	Type string `json:"type"`
}

type anthropicOutput struct {
	Effort string `json:"effort"`
}

func buildAnthropicPayload(model, effort string, messages []Message, toolsSchema json.RawMessage, sendThinking bool) ([]byte, error) {
	system, anthropicMsgs := convertHistory(messages)
	tools, err := convertTools(toolsSchema)
	if err != nil {
		return nil, errors.New("cannot encode tools schema")
	}
	p := anthropicPayload{
		Model:     model,
		MaxTokens: 32000,
		System:    system,
		Messages:  anthropicMsgs,
		Tools:     tools,
		Stream:    true,
	}
	if sendThinking && effort != "" {
		p.Thinking = &anthropicThink{Type: "adaptive"}
		p.OutputConfig = &anthropicOutput{Effort: effort}
	}
	data, err := json.Marshal(p)
	if err != nil {
		return nil, errors.New("cannot encode chat request")
	}
	return data, nil
}
