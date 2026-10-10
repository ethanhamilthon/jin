package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"jin/internal/session"
)

func (c *Client) Command(ctx context.Context, command Command, value any) error {
	limit := 30 * time.Second
	if command.Action == "title" {
		limit = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	if command.ID == "" {
		command.ID = session.NewID()
	}
	command.Version = c.Version
	body, err := json.Marshal(command)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "http://daemon/command", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Transport: c.http.Transport}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var result Result
	if err := json.NewDecoder(io.LimitReader(response.Body, 32<<20)).Decode(&result); err != nil {
		return err
	}
	if result.Error != "" {
		return errors.New(result.Error)
	}
	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}
	if value == nil {
		return nil
	}
	return json.Unmarshal(result.Value, value)
}

func (c *Client) Create(ctx context.Context, path string) (session.Snapshot, error) {
	var snap session.Snapshot
	err := c.Command(ctx, Command{Action: "create", Path: path}, &snap)
	return snap, err
}
func (c *Client) Open(ctx context.Context, id string) (session.Snapshot, error) {
	var snap session.Snapshot
	err := c.Command(ctx, Command{Action: "open", Session: id}, &snap)
	return snap, err
}
func (c *Client) Snapshot(ctx context.Context, id string) (session.Snapshot, error) {
	var snap session.Snapshot
	err := c.Command(ctx, Command{Action: "snapshot", Session: id}, &snap)
	return snap, err
}
