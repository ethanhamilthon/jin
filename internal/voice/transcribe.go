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

// Client talks to an OpenAI-compatible endpoint or a fal.run speech model.
type Client struct {
	BaseURL, APIKey, Model, Language string
	HTTP                             *http.Client
}

func (c Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 90 * time.Second}
}

// Transcribe sends PCM audio and returns the text. A fal.run endpoint gets
// fal's own JSON request; any other one gets the OpenAI multipart form.
func (c Client) Transcribe(ctx context.Context, pcm []byte) (string, error) {
	req, err := c.request(ctx, pcm)
	if err != nil {
		return "", err
	}
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
		Text *string `json:"text"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("transcription reply is not JSON: %w", err)
	}
	if out.Text == nil {
		return "", fmt.Errorf("transcription reply has no text field; check the endpoint")
	}
	return strings.TrimSpace(*out.Text), nil
}

// Check sends a second of silence to see that the endpoint, key and model
// work. The reply text does not matter.
func (c Client) Check(ctx context.Context) error {
	_, err := c.Transcribe(ctx, make([]byte, bytesPerSecond))
	return err
}

func (c Client) request(ctx context.Context, pcm []byte) (*http.Request, error) {
	if isFal(c.BaseURL) {
		return c.falRequest(ctx, pcm)
	}
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	file, err := form.CreateFormFile("file", "voice.wav")
	if err != nil {
		return nil, err
	}
	if _, err := file.Write(WAV(pcm)); err != nil {
		return nil, err
	}
	fields := map[string]string{"model": c.Model, "response_format": "json", "temperature": "0"}
	if c.Language != "" {
		fields["language"] = c.Language
	}
	for name, value := range fields {
		if err := form.WriteField(name, value); err != nil {
			return nil, err
		}
	}
	if err := form.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/audio/transcriptions", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", form.FormDataContentType())
	return req, nil
}
