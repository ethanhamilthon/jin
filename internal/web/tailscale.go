package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const macAppCLI = "/Applications/Tailscale.app/Contents/MacOS/Tailscale"

// tailscale runs the tailscale command and returns its combined output.
var tailscale = func(ctx context.Context, args ...string) ([]byte, error) {
	bin, err := exec.LookPath("tailscale")
	if err != nil {
		if _, statErr := os.Stat(macAppCLI); statErr != nil {
			return nil, errors.New("Tailscale is not installed. Install it from https://tailscale.com/download and sign in")
		}
		bin = macAppCLI
	}
	return exec.CommandContext(ctx, bin, args...).CombinedOutput()
}

type tailscaleStatus struct {
	BackendState string
	Self         struct{ DNSName string }
	CertDomains  []string
}

// tailnetName returns this computer's HTTPS name on the tailnet, or an
// error that says what the user must do.
func tailnetName(ctx context.Context) (string, error) {
	out, err := tailscale(ctx, "status", "--json")
	if err != nil && len(out) == 0 {
		return "", err
	}
	var st tailscaleStatus
	if json.Unmarshal(out, &st) != nil {
		return "", fmt.Errorf("tailscale status failed: %s", strings.TrimSpace(string(out)))
	}
	if st.BackendState != "Running" {
		return "", fmt.Errorf("Tailscale is not connected (state %s). Sign in or run `tailscale up`", st.BackendState)
	}
	name := strings.TrimSuffix(st.Self.DNSName, ".")
	if name == "" || len(st.CertDomains) == 0 {
		return "", errors.New("HTTPS is not enabled in your tailnet. Turn on MagicDNS and HTTPS Certificates at https://login.tailscale.com/admin/dns")
	}
	return name, nil
}

func serveArgs(port string) []string {
	return []string{"serve", "--bg", "--https=443", "http://127.0.0.1:" + port}
}

// tailnetServe publishes the local port over HTTPS inside the tailnet.
func tailnetServe(ctx context.Context, port string) error {
	if out, err := tailscale(ctx, serveArgs(port)...); err != nil {
		return fmt.Errorf("tailscale serve failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func tailnetStop() {
	_, _ = tailscale(context.Background(), "serve", "--https=443", "off")
}
