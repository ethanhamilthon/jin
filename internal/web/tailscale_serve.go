//go:build unix

package web

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var errForeignEntry = errors.New("Tailscale HTTPS 443 is already configured; jin will not replace another Serve entry")

// proxyURL is where jin serves the page: a Serve entry with this exact target
// is jin's own, left by this run or by a run that crashed.
func proxyURL(port string) string { return "http://127.0.0.1:" + port }

func localServeEntry(ctx context.Context) (proxy string, occupied bool, err error) {
	status, err := tailnetServeStatus(ctx)
	if err != nil {
		return "", false, err
	}
	proxy, occupied = status.entry443()
	return proxy, occupied, nil
}

// tailnetAvailable refuses when HTTPS 443 carries a Serve entry that jin did
// not make. An entry that already forwards to this port is jin's own, so jin
// takes it over instead of refusing to start.
func tailnetAvailable(ctx context.Context, port string) error {
	proxy, occupied, err := localServeEntry(ctx)
	if err != nil {
		return err
	}
	switch {
	case !occupied:
		return nil
	case proxy == proxyURL(port):
		return nil
	case proxy == "":
		return errForeignEntry
	default:
		return errors.New("Tailscale HTTPS 443 already forwards to " + proxy + "; jin will not replace another Serve entry")
	}
}

func tailnetDisable(ctx context.Context) error {
	if out, err := tailscale(ctx, "serve", "--https=443", "off"); err != nil {
		return fmt.Errorf("tailscale serve stop failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}
