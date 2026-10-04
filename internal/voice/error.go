package voice

import (
	"encoding/json"
	"strings"
)

func errorText(data []byte) string {
	var out struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		Detail any `json:"detail"`
	}
	if json.Unmarshal(data, &out) == nil {
		if out.Error.Message != "" {
			return out.Error.Message
		}
		if detail, ok := out.Detail.(string); ok && detail != "" {
			return detail
		}
	}
	text := strings.TrimSpace(string(data))
	if len(text) > 200 {
		text = text[:200]
	}
	return text
}
