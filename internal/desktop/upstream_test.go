//go:build desktop

package desktop

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const token = "0123456789abcdef0123456789abcdef0123456789abcdef"

func TestParseUpstream(t *testing.T) {
	for _, raw := range []string{"http://127.0.0.1:7400/?token=" + token, "http://localhost:7400/?token=" + token} {
		if _, err := parseUpstream(raw); err != nil {
			t.Errorf("rejected %q: %v", raw, err)
		}
	}
	for _, raw := range []string{
		"https://127.0.0.1:7400/?token=" + token, "http://evil.test:7400/?token=" + token,
		"http://127.0.0.1:7400/x?token=" + token, "http://127.0.0.1:7400/?token=short",
		"http://127.0.0.1:7400/?token=" + token + "&target=x", "http://u@127.0.0.1:7400/?token=" + token,
	} {
		if _, err := parseUpstream(raw); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
}

func TestProxyRefusesForeignOriginAndHidesSecrets(t *testing.T) {
	var seen *http.Request
	daemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = r }))
	defer daemon.Close()
	p := &proxy{}
	base, _ := parseUpstream("http://" + daemon.Listener.Addr().String() + "/?token=" + token)
	p.to.Store(&link{base: base, cookie: "device", client: daemon.Client()})
	req := httptest.NewRequest("GET", "/api/state?token=leak&x=1", nil)
	req.Header.Set("Origin", "https://evil.test")
	w := httptest.NewRecorder()
	p.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden || seen != nil {
		t.Fatalf("foreign origin: code=%d forwarded=%v", w.Code, seen != nil)
	}
	req = httptest.NewRequest("GET", "/api/state?token=leak&x=1", nil)
	req.Header.Set("Authorization", "Bearer secret")
	w = httptest.NewRecorder()
	p.ServeHTTP(w, req)
	if seen == nil || seen.URL.Query().Has("token") || seen.URL.Query().Get("x") != "1" || seen.Header.Get("Authorization") != "" ||
		seen.Header.Get("Cookie") != "jin_device=device" || seen.Header.Get("Origin") != "http://"+daemon.Listener.Addr().String() {
		t.Fatalf("bad upstream request: %+v", seen)
	}
}

func TestProxyStreamsEventsWithoutBuffering(t *testing.T) {
	release := make(chan struct{})
	daemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: one\n\n"))
		w.(http.Flusher).Flush()
		<-release
	}))
	defer daemon.Close()
	defer close(release)
	p := &proxy{}
	base, _ := parseUpstream("http://" + daemon.Listener.Addr().String() + "/?token=" + token)
	p.to.Store(&link{base: base, cookie: "d", client: daemon.Client()})
	front := httptest.NewServer(p)
	defer front.Close()
	resp, err := http.Get(front.URL + "/api/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	got := make(chan string, 1)
	go func() {
		buf := make([]byte, 64)
		n, _ := resp.Body.Read(buf)
		got <- string(buf[:n])
	}()
	select {
	case text := <-got:
		if !strings.Contains(text, "data: one") {
			t.Fatalf("event = %q", text)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("event was buffered")
	}
}

func TestProxyReportsStartupState(t *testing.T) {
	p := &proxy{}
	w := httptest.NewRecorder()
	p.ServeHTTP(w, httptest.NewRequest("GET", "/api/state", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("code=%d", w.Code)
	}
}
