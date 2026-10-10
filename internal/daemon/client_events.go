package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"jin/internal/session"
)

func (c *Client) Events(ctx context.Context) <-chan session.Event {
	events := make(chan session.Event, 256)
	go func() {
		defer close(events)
		for ctx.Err() == nil {
			_ = c.readEvents(ctx, events)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}()
	return events
}

func (c *Client) readEvents(ctx context.Context, events chan<- session.Event) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://daemon/events?version="+url.QueryEscape(c.Version), nil)
	if err != nil {
		return err
	}
	client := &http.Client{Transport: c.http.Transport}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil
	}
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var event session.Event
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
			return err
		}
		select {
		case events <- event:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return scanner.Err()
}
