package daemon

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"jin/internal/session"
)

func TestEventStreamVersionAndResync(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager := session.NewManager(ctx, nil, "v1", nil)
	server := httptest.NewServer(routes("v1", manager, func() {}))
	defer server.Close()
	response, err := http.Get(server.URL + "/events?version=v2")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("status: %d", response.StatusCode)
	}
	response, err = http.Get(server.URL + "/events?version=v1")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	scanner := bufio.NewScanner(response.Body)
	if !scanner.Scan() || scanner.Text() != `data: {"type":"resync"}` {
		t.Fatalf("initial event: %s", scanner.Text())
	}
}
