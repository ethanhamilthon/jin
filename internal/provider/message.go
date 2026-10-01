package provider

import (
	"encoding/json"
	"strconv"
	"strings"
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

type contentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL string `json:"url"`
}

// MarshalJSON writes the API shape, which is also the stored shape: a plain
// string content, or a list of parts when the message carries images.
func (m Message) MarshalJSON() ([]byte, error) {
	type plain Message
	if len(m.Images) == 0 {
		return json.Marshal(plain(m))
	}
	parts := []contentPart{}
	if m.Content != "" {
		parts = append(parts, contentPart{Type: "text", Text: m.Content})
	}
	for _, image := range m.Images {
		parts = append(parts, contentPart{Type: "image_url", ImageURL: &imageURL{URL: image.url()}})
	}
	return json.Marshal(struct {
		plain
		Content []contentPart `json:"content"`
	}{plain(m), parts})
}

func (m *Message) UnmarshalJSON(data []byte) error {
	type plain Message
	aux := struct {
		*plain
		Content json.RawMessage `json:"content"`
	}{plain: (*plain)(m)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	m.Content, m.Images = "", nil
	var text string
	if json.Unmarshal(aux.Content, &text) == nil {
		m.Content = text
		return nil
	}
	var parts []contentPart
	if json.Unmarshal(aux.Content, &parts) != nil {
		return nil
	}
	var texts []string
	for _, part := range parts {
		if part.Type == "text" {
			texts = append(texts, part.Text)
		} else if part.ImageURL != nil {
			m.Images = append(m.Images, parseDataURL(part.ImageURL.URL))
		}
	}
	m.Content = strings.Join(texts, "\n")
	return nil
}

func parseDataURL(url string) Image {
	header, data, _ := strings.Cut(strings.TrimPrefix(url, "data:"), ",")
	return Image{MimeType: strings.TrimSuffix(header, ";base64"), Data: data}
}
