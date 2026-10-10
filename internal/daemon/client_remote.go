package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type RemoteCommand struct {
	Version string
	Action  string
	Path    string
	ID      string
	Name    string
	Enabled bool
}

func (c *Client) Remote(ctx context.Context, command RemoteCommand, value any) error {
	command.Version = c.Version
	body, err := json.Marshal(command)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "http://daemon/remote", bytes.NewReader(body))
	if err != nil {
		return err
	}
	client := &http.Client{Transport: c.http.Transport}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var result Result
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return err
	}
	if result.Error != "" {
		return errors.New(result.Error)
	}
	if value == nil {
		return nil
	}
	return json.Unmarshal(result.Value, value)
}
