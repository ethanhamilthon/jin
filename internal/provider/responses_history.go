package provider

import "encoding/json"

func responsesHistory(messages []Message) []json.RawMessage {
	out := make([]json.RawMessage, 0, len(messages))
	appendItem := func(item any) {
		data, _ := json.Marshal(item)
		out = append(out, data)
	}
	for _, m := range messages {
		if m.Role == "assistant" && m.Native != nil && m.Native.Kind == KindResponses {
			out = append(out, m.Native.Items...)
			continue
		}
		if m.Role == "tool" {
			appendItem(map[string]any{
				"type": "function_call_output", "call_id": m.ToolCallID, "output": m.Content,
			})
			if len(m.Images) > 0 {
				appendItem(responsesMessage("user", "", m.Images))
			}
			continue
		}
		role := m.Role
		if role == "system" {
			role = "developer"
		}
		if m.Content != "" || len(m.Images) > 0 {
			appendItem(responsesMessage(role, m.Content, m.Images))
		}
		for _, call := range m.ToolCalls {
			appendItem(map[string]any{"type": "function_call", "call_id": call.ID,
				"name": call.Function.Name, "arguments": call.Function.Arguments})
		}
	}
	return out
}

func responsesMessage(role, text string, images []Image) map[string]any {
	parts := make([]map[string]any, 0, len(images)+1)
	if text != "" {
		kind := "input_text"
		if role == "assistant" {
			kind = "output_text"
		}
		parts = append(parts, map[string]any{"type": kind, "text": text})
	}
	for _, image := range images {
		parts = append(parts, map[string]any{"type": "input_image", "image_url": image.url()})
	}
	return map[string]any{"type": "message", "role": role, "content": parts}
}
