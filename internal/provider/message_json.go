package provider

import (
	"encoding/json"
	"strings"
)

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
