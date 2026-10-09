package web

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func fakeTailscale(t *testing.T, out string, err error) {
	old := tailscale
	tailscale = func(context.Context, ...string) ([]byte, error) { return []byte(out), err }
	t.Cleanup(func() { tailscale = old })
}

func TestTailnetName(t *testing.T) {
	cases := []struct{ name, out, want, fail string }{
		{"ready", `{"BackendState":"Running","Self":{"DNSName":"mac.tail.ts.net."},"CertDomains":["mac.tail.ts.net"]}`, "mac.tail.ts.net", ""},
		{"signed out", `{"BackendState":"NeedsLogin"}`, "", "not connected"},
		{"no https", `{"BackendState":"Running","Self":{"DNSName":"mac.tail.ts.net."}}`, "", "HTTPS is not enabled"},
		{"garbage", `not json`, "", "tailscale status failed"},
	}
	for _, c := range cases {
		fakeTailscale(t, c.out, nil)
		got, err := tailnetName(context.Background())
		if c.fail == "" && (err != nil || got != c.want) {
			t.Errorf("%s: got %q, %v", c.name, got, err)
		}
		if c.fail != "" && (err == nil || !strings.Contains(err.Error(), c.fail)) {
			t.Errorf("%s: error %v, want %q", c.name, err, c.fail)
		}
	}
}

func TestTailnetMissing(t *testing.T) {
	fakeTailscale(t, "", errors.New("Tailscale is not installed"))
	if _, err := tailnetName(context.Background()); err == nil {
		t.Fatal("a missing tailscale must be an error")
	}
}

func TestTailnetServeError(t *testing.T) {
	fakeTailscale(t, "Serve is not enabled on your tailnet", errors.New("exit 1"))
	err := tailnetServe(context.Background(), "7373")
	if err == nil || !strings.Contains(err.Error(), "Serve is not enabled") {
		t.Fatalf("got %v", err)
	}
}
