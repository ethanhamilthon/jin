package provider

import (
	"encoding/json"
	"errors"
)

type responsesTool struct {
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"`
	Strict      bool            `json:"strict"`
}

func responsesTools(schema json.RawMessage) ([]responsesTool, error) {
	if len(schema) == 0 {
		return nil, nil
	}
	var raw []struct {
		Type     string `json:"type"`
		Function struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			Parameters  json.RawMessage `json:"parameters"`
			Strict      bool            `json:"strict"`
		} `json:"function"`
	}
	if err := json.Unmarshal(schema, &raw); err != nil {
		return nil, errors.New("cannot encode tools schema")
	}
	tools := make([]responsesTool, 0, len(raw))
	for _, tool := range raw {
		if tool.Type != "function" || tool.Function.Name == "" || len(tool.Function.Parameters) == 0 {
			return nil, errors.New("unsupported Responses tool schema")
		}
		tools = append(tools, responsesTool{Type: "function", Name: tool.Function.Name,
			Description: tool.Function.Description, Parameters: tool.Function.Parameters, Strict: tool.Function.Strict})
	}
	return tools, nil
}
