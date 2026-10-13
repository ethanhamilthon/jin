package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRetiredCLIProxyRoutesReturnUnavailable(t *testing.T) {
	mux := http.NewServeMux()
	(&server{}).cliproxyRoutes(mux)
	server := httptest.NewServer(mux)
	defer server.Close()
	for _, route := range []struct{ method, path string }{
		{"GET", "/api/cliproxy"},
		{"GET", "/api/cliproxy/versions"},
		{"POST", "/api/cliproxy/install"},
		{"POST", "/api/cliproxy/update"},
		{"GET", "/api/cliproxy/accounts"},
		{"POST", "/api/cliproxy/providers"},
		{"POST", "/api/cliproxy/login"},
		{"POST", "/api/cliproxy/poll"},
		{"POST", "/api/cliproxy/cancel"},
		{"POST", "/api/cliproxy/logout"},
	} {
		if status := call(t, server, route.method, route.path, `{}`, nil); status != 503 {
			t.Errorf("%s %s status=%d, want 503", route.method, route.path, status)
		}
	}
}
