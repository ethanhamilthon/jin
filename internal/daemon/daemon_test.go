package daemon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"jin/internal/session"
)

func TestVersionAndStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager := session.NewManager(ctx, nil, "v1", func(session.Event) {})
	stopped := false
	server := httptest.NewServer(routes("v1", manager, func() { stopped = true }))
	defer server.Close()
	client := &Client{http: &http.Client{Transport: rewriteTransport{server.URL}}}
	if err := client.Check(ctx, "v2"); err == nil || !strings.Contains(err.Error(), "differs") {
		t.Fatalf("version error: %v", err)
	}
	if err := client.Stop(ctx, "v2", true); err == nil {
		t.Fatal("different client version stopped daemon")
	}
	if stopped {
		t.Fatal("daemon stopped on version mismatch")
	}
	if err := client.Stop(ctx, "v1", false); err != nil {
		t.Fatal(err)
	}
	if !stopped {
		t.Fatal("daemon did not stop")
	}
}

type rewriteTransport struct{ url string }

func (t rewriteTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	req, err := http.NewRequestWithContext(r.Context(), r.Method, t.url+r.URL.RequestURI(), r.Body)
	if err != nil {
		return nil, err
	}
	return http.DefaultTransport.RoundTrip(req)
}

func TestCommandValidation(t *testing.T) {
	for _, args := range [][]string{nil, {"unknown"}, {"start", "--force"}, {"stop", "extra"}} {
		if err := Main(context.Background(), args, "v1", t.TempDir(), &strings.Builder{}); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestSocketIsolation(t *testing.T) {
	root := t.TempDir()
	listener, cleanup, err := listen(root)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if _, _, err := listen(root); err == nil {
		t.Fatal("second daemon acquired lock")
	}
	info, err := os.Stat(socketPath(root))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("socket permissions: %v", info.Mode())
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reply(w, 200, Status{Version: "v1"}) })}
	go server.Serve(listener)
	defer server.Close()
	if err := NewClient(root).Check(context.Background(), "v1"); err != nil {
		t.Fatal(err)
	}
}
