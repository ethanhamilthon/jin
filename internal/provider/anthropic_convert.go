package provider

import (
	"bytes"
	"encoding/json"
	"strings"
)

type anthropicMsg struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type anthropicTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicImageBlock struct {
	Type   string `json:"type"`
	Source struct {
		Type      string `json:"type"`
		MediaType string `json:"media_type"`
		Data      string `json:"data"`
	} `json:"source"`
}

func convertHistory(messages []Message) (string, []anthropicMsg) {
	var systemParts []string
	var out []anthropicMsg
	for i := 0; i < len(messages); {
		switch m := messages[i]; m.Role {
		case "system":
			if strings.TrimSpace(m.Content) != "" {
				systemParts = append(systemParts, m.Content)
			}
			i++
		case "tool":
			var blocks []any
			for i < len(messages) && messages[i].Role == "tool" {
				blocks = append(blocks, anthropicToolResultBlock{
					Type: "tool_result", ToolUseID: messages[i].ToolCallID, Content: messages[i].Content,
				})
				i++
			}
			out = append(out, anthropicMsg{Role: "user", Content: blocks})
		case "user":
			out = append(out, convertUserMessage(m))
			i++
		case "assistant":
			out = append(out, convertAssistantMessage(m))
			i++
		default:
			i++
		}
	}
	return strings.Join(systemParts, "\n\n"), out
}

func convertUserMessage(m Message) anthropicMsg {
	if len(m.Images) == 0 {
		return anthropicMsg{Role: "user", Content: m.Content}
	}
	var blocks []any
	if m.Content != "" {
		blocks = append(blocks, anthropicTextBlock{Type: "text", Text: m.Content})
	}
	for _, img := range m.Images {
		var b anthropicImageBlock
		b.Type, b.Source.Type, b.Source.MediaType, b.Source.Data = "image", "base64", img.MimeType, img.Data
		blocks = append(blocks, b)
	}
	return anthropicMsg{Role: "user", Content: blocks}
}

func convertAssistantMessage(m Message) anthropicMsg {
	if len(m.ToolCalls) == 0 {
		return anthropicMsg{Role: "assistant", Content: m.Content}
	}
	var blocks []any
	if m.Content != "" {
		blocks = append(blocks, anthropicTextBlock{Type: "text", Text: m.Content})
	}
	for _, tc := range m.ToolCalls {
		raw := json.RawMessage(tc.Function.Arguments)
		if len(bytes.TrimSpace(raw)) == 0 || !json.Valid(raw) {
			raw = json.RawMessage("{}")
		}
		blocks = append(blocks, anthropicToolUseBlock{
			Type: "tool_use", ID: tc.ID, Name: tc.Function.Name, Input: raw,
		})
	}
	return anthropicMsg{Role: "assistant", Content: blocks}
}
