package web

import (
	"context"
	"errors"
	"net"
	"testing"

	"jin/internal/store"
)

func TestRemoteToggleKeepsDevicesAndInvalidatesCodes(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := &server{ctx: ctx, db: db, hub: newHub(), guard: newGuard("1234", nil, db)}
	service := &Service{server: s, listener: listener}
	remote := &remoteAccess{service: service}
	old := tailscale
	defer func() { tailscale = old }()
	tailscale = func(_ context.Context, args ...string) ([]byte, error) {
		if args[0] == "status" {
			return []byte(`{"BackendState":"Running","Self":{"DNSName":"mac.tail.net."},"CertDomains":["mac.tail.net"]}`), nil
		}
		if len(args) > 1 && args[1] == "status" {
			return []byte(`{}`), nil
		}
		return nil, nil
	}
	if err := remote.set(ctx, true); err != nil {
		t.Fatal(err)
	}
	code, _ := s.guard.codes.issue()
	if !s.guard.allowed("mac.tail.net") {
		t.Fatal("remote host not admitted")
	}
	if err := remote.set(ctx, false); err != nil {
		t.Fatal(err)
	}
	if s.guard.allowed("mac.tail.net") || s.guard.codes.take(code) {
		t.Fatal("remote access or old code remains valid")
	}
	saved, _ := db.Setting(remoteKey)
	if saved != "false" {
		t.Fatalf("saved: %s", saved)
	}
}

func TestOccupiedServeIsNotReplaced(t *testing.T) {
	fakeTailscale(t, `{"TCP":{"443":{"HTTPS":true}}}`, nil)
	if err := tailnetAvailable(context.Background()); err == nil {
		t.Fatal("accepted occupied HTTPS port")
	}
	fakeTailscale(t, "failed", errors.New("exit 1"))
	if err := tailnetAvailable(context.Background()); err == nil {
		t.Fatal("ignored status error")
	}
}
