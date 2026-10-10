package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"jin/internal/session"
)

type Status struct {
	Version string `json:"version"`
	PID     int    `json:"pid"`
	Working bool   `json:"working"`
	Tasks   int    `json:"tasks"`
}

type Client struct {
	http    *http.Client
	Version string
	// ID names this client in events that only its initiator acts on,
	// such as a handoff: every client gets the event, one client follows it.
	ID string
}

func NewClient(root string) *Client {
	return &Client{ID: session.NewID(), http: &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socketPath(root))
		},
	}, Timeout: 5 * time.Second}}
}

func (c *Client) call(ctx context.Context, method, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, method, "http://daemon"+path, nil)
	if err != nil {
		return err
	}
	response, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		var failure struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(response.Body).Decode(&failure); err != nil {
			return err
		}
		return fmt.Errorf("%s", failure.Error)
	}
	if result == nil {
		return nil
	}
	return json.NewDecoder(response.Body).Decode(result)
}

func (c *Client) Status(ctx context.Context) (Status, error) {
	var status Status
	err := c.call(ctx, http.MethodGet, "/status", &status)
	return status, err
}

func (c *Client) Check(ctx context.Context, version string) error {
	status, err := c.Status(ctx)
	if err != nil {
		return err
	}
	if status.Version != version {
		return fmt.Errorf("client version %s differs from daemon version %s; stop the daemon and start it with this binary", version, status.Version)
	}
	c.Version = version
	return nil
}

func (c *Client) Stop(ctx context.Context, version string, force bool) error {
	if err := c.Check(ctx, version); err != nil {
		return err
	}
	return c.call(ctx, http.MethodPost, "/stop?force="+fmt.Sprint(force)+"&version="+url.QueryEscape(version), nil)
}
