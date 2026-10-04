package voice

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

// isFal reports whether the base URL is a fal.run host. fal speaks its own
// JSON, not the OpenAI form, and has no /audio/transcriptions.
func isFal(baseURL string) bool {
	u, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host == "fal.run" || strings.HasSuffix(host, ".fal.run")
}

// falEndpoint is the synchronous URL of the fal app: the path of the base
// URL when it has one, otherwise the model, like fal-ai/wizper.
func falEndpoint(baseURL, model string) string {
	path := ""
	if u, err := url.Parse(baseURL); err == nil {
		path = strings.Trim(u.Path, "/")
	}
	if path == "" {
		path = strings.Trim(model, "/")
	}
	if !strings.Contains(path, "/") {
		path = "fal-ai/" + path
	}
	return "https://fal.run/" + path
}

func (c Client) falRequest(ctx context.Context, pcm []byte) (*http.Request, error) {
	body := map[string]any{
		"audio_url": "data:audio/x-wav;base64," + base64.StdEncoding.EncodeToString(WAV(pcm)),
		"task":      "transcribe",
		"language":  nil,
	}
	if c.Language != "" {
		body["language"] = c.Language
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, falEndpoint(c.BaseURL, c.Model), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Key "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
