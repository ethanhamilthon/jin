package provider

import (
	"strconv"
)

type Message struct {
	Role             string     `json:"role"`
	Content          string     `json:"content,omitempty"`
	Images           []Image    `json:"-"`
	ReasoningContent string     `json:"reasoning_content,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID       string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// Image is a base64 encoded picture attached to a message.
type Image struct {
	MimeType string
	Data     string
}

func (i Image) url() string { return "data:" + i.MimeType + ";base64," + i.Data }

// Label is a short description for the chat, such as "image/png · 84 KB".
func (i Image) Label() string {
	kb := (len(i.Data)*3/4 + 1023) / 1024
	size := strconv.Itoa(kb) + " KB"
	if kb >= 1024 {
		size = strconv.FormatFloat(float64(kb)/1024, 'f', 1, 64) + " MB"
	}
	return i.MimeType + " · " + size
}
