package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Client talks to an OpenAI-compatible /audio/transcriptions endpoint.
type Client struct {
	BaseURL, APIKey, Model, Language string
	HTTP                             *http.Client
}

func (c Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// Transcribe sends PCM audio and returns the text.
func (c Client) Transcribe(ctx context.Context, pcm []byte) (string, error) {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	file, err := form.CreateFormFile("file", "voice.wav")
	if err != nil {
		return "", err
	}
	if _, err := file.Write(WAV(pcm)); err != nil {
		return "", err
	}
	fields := map[string]string{"model": c.Model, "response_format": "json", "temperature": "0"}
	if c.Language != "" {
		fields["language"] = c.Language
	}
	for name, value := range fields {
		if err := form.WriteField(name, value); err != nil {
			return "", err
		}
	}
	if err := form.Close(); err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/audio/transcriptions", &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", form.FormDataContentType())
	resp, err := c.http().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("transcription failed: %s: %s", resp.Status, errorText(data))
	}
	var out struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("transcription reply is not JSON: %w", err)
	}
	return strings.TrimSpace(out.Text), nil
}

func errorText(data []byte) string {
	var out struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(data, &out) == nil && out.Error.Message != "" {
		return out.Error.Message
	}
	text := strings.TrimSpace(string(data))
	if len(text) > 200 {
		text = text[:200]
	}
	return text
}
