package provider

import (
	"encoding/json"
	"fmt"
)

func responsesOutput(items []json.RawMessage, model string) (Message, error) {
	msg := Message{Role: "assistant", Native: &NativeMessage{Kind: KindResponses, Model: model, Items: items}}
	for _, raw := range items {
		var item struct {
			Type      string `json:"type"`
			CallID    string `json:"call_id"`
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
			Content   []struct {
				Type    string `json:"type"`
				Text    string `json:"text"`
				Refusal string `json:"refusal"`
			} `json:"content"`
			Summary []struct {
				Text string `json:"text"`
			} `json:"summary"`
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			return Message{}, fmt.Errorf("invalid Responses output item: %w", err)
		}
		switch item.Type {
		case "message":
			for _, part := range item.Content {
				msg.Content += part.Text + part.Refusal
			}
		case "reasoning":
			for _, part := range item.Summary {
				msg.ReasoningContent += part.Text
			}
		case "function_call":
			if item.CallID == "" || item.Name == "" || !json.Valid([]byte(item.Arguments)) {
				return Message{}, fmt.Errorf("invalid Responses function call")
			}
			call := ToolCall{ID: item.CallID, Type: "function"}
			call.Function.Name, call.Function.Arguments = item.Name, item.Arguments
			msg.ToolCalls = append(msg.ToolCalls, call)
		default:
			return Message{}, fmt.Errorf("unsupported Responses output item %q", item.Type)
		}
	}
	return msg, nil
}
