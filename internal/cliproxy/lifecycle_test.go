package cliproxy

import (
	"context"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func waitSocketClosed(t *testing.T, root string) {
	t.Helper()
	deadline := time.Now().Add(8 * time.Second)
	for {
		conn, err := net.DialTimeout("unix", socketPath(root), 50*time.Millisecond)
		if err != nil {
			return
		}
		conn.Close()
		if time.Now().After(deadline) {
			t.Fatal("proxy supervisor did not stop")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestSharedProxyStopsOnlyAfterClientsAndRequests(t *testing.T) {
	root := t.TempDir()
	installFixture(t, root, DefaultVersion)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	one, two := NewClient(root), NewClient(root)
	defer one.Close()
	defer two.Close()
	endpoint, key, release, err := one.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	endpointTwo, keyTwo, releaseTwo, err := two.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	releaseTwo()
	if endpoint != endpointTwo || key != keyTwo {
		t.Fatal("proxy was not shared")
	}
	one.Close()
	two.Close()
	time.Sleep(150 * time.Millisecond)
	if _, err := os.Stat(socketPath(root)); err != nil {
		t.Fatal("active request lost its proxy")
	}
	release()
	waitSocketClosed(t, root)
}

func TestAllLoginSourcesCanStartPollAndCancel(t *testing.T) {
	root := t.TempDir()
	installFixture(t, root, DefaultVersion)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	client := NewClient(root)
	defer client.Close()
	for _, profile := range []string{"claude", "codex", "antigravity"} {
		login, err := client.Login(ctx, profile)
		if err != nil || login.State == "" || !strings.HasPrefix(login.URL, "https://") {
			t.Fatalf("%s login failed: %v", profile, err)
		}
		pending, err := client.Poll(ctx, profile, login.State)
		if err != nil || pending.Status != "wait" {
			t.Fatalf("%s poll: %v", profile, err)
		}
		if err = client.Cancel(ctx, login.State); err != nil {
			t.Fatal(err)
		}
	}
	client.Close()
	waitSocketClosed(t, root)
}
