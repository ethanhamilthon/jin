package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"jin/internal/web"
)

func (c *Client) Web(ctx context.Context, path string, options web.Options) (string, error) {
	body, err := json.Marshal(struct {
		Version, Path string
		Options       web.Options
	}{c.Version, path, options})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "http://daemon/web", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	response, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	var value struct{ URL, Error string }
	if err := json.NewDecoder(response.Body).Decode(&value); err != nil {
		return "", err
	}
	if value.Error != "" {
		return "", errors.New(value.Error)
	}
	return value.URL, nil
}
