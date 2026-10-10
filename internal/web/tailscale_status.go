package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// serveStatus is the part of `tailscale serve status --json` that jin reads.
type serveStatus struct {
	TCP map[string]json.RawMessage `json:"TCP"`
	Web map[string]serveHost       `json:"Web"`
}

type serveHost struct {
	Handlers map[string]serveHandler `json:"Handlers"`
}

type serveHandler struct {
	Proxy string `json:"Proxy"`
}

// entry443 says whether HTTPS 443 is configured and, when it is, where it
// forwards to: the proxy address names the entry jin may replace.
func (s serveStatus) entry443() (proxy string, occupied bool) {
	_, occupied = s.TCP["443"]
	for host, entry := range s.Web {
		if !strings.HasSuffix(host, ":443") {
			continue
		}
		occupied = true
		for _, handler := range entry.Handlers {
			proxy = handler.Proxy
		}
	}
	return proxy, occupied
}

func tailnetServeStatus(ctx context.Context) (serveStatus, error) {
	out, err := tailscale(ctx, "serve", "status", "--json")
	if err != nil {
		return serveStatus{}, fmt.Errorf("tailscale serve status failed: %s", strings.TrimSpace(string(out)))
	}
	var status serveStatus
	if err := json.Unmarshal(out, &status); err != nil {
		return serveStatus{}, errors.New("could not read Tailscale Serve configuration")
	}
	return status, nil
}
