package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func tailnetAvailable(ctx context.Context) error {
	out, err := tailscale(ctx, "serve", "status", "--json")
	if err != nil {
		return fmt.Errorf("tailscale serve status failed: %s", strings.TrimSpace(string(out)))
	}
	var status struct {
		TCP map[string]json.RawMessage
		Web map[string]json.RawMessage
	}
	if err := json.Unmarshal(out, &status); err != nil {
		return errors.New("could not read Tailscale Serve configuration")
	}
	if _, occupied := status.TCP["443"]; occupied {
		return errors.New("Tailscale HTTPS 443 is already configured; jin will not replace another Serve entry")
	}
	for host := range status.Web {
		if strings.HasSuffix(host, ":443") {
			return errors.New("Tailscale HTTPS 443 is already configured; jin will not replace another Serve entry")
		}
	}
	return nil
}

func tailnetDisable(ctx context.Context) error {
	if out, err := tailscale(ctx, "serve", "--https=443", "off"); err != nil {
		return fmt.Errorf("tailscale serve stop failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}
