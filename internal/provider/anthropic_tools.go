package provider

import (
	"bytes"
	"encoding/json"
)

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
