package provider

import (
	"bytes"
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

func convertTools(toolsSchema json.RawMessage) ([]anthropicTool, error) {
	if len(bytes.TrimSpace(toolsSchema)) == 0 {
		return nil, nil
	}
	var raw []struct {
		Function struct {
			Name        string          `json:"name"`
			Description string          `json:"description,omitempty"`
			Parameters  json.RawMessage `json:"parameters,omitempty"`
		} `json:"function"`
	}
	if err := json.Unmarshal(toolsSchema, &raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	tools := make([]anthropicTool, len(raw))
	for i, t := range raw {
		params := t.Function.Parameters
		if len(bytes.TrimSpace(params)) == 0 {
			params = json.RawMessage(`{"type":"object"}`)
		}
		tools[i] = anthropicTool{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: params,
		}
	}
	return tools, nil
}
