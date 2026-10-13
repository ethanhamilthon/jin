package desktop

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterAllowsOnlyKnownRoutes(t *testing.T) {
	ok := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	router := newRouter(ok)
	for _, tc := range []struct {
		method, path string
		want         int
	}{
		{"GET", "/api/state", 200}, {"GET", "/api/events", 200}, {"POST", "/api/sessions/abc/send", 200},
		{"GET", "/api/sessions/abc/entries/2/image", 200}, {"DELETE", "/api/devices/x", 200},
		{"POST", "/api/admin", 404}, {"PUT", "/api/state", 405}, {"GET", "/api/cliproxy", 404},
		{"POST", "/api/sessions/abc/generate-title", 404}, {"GET", "/api/../x", 404}, {"GET", "/api//state", 404},
	} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.want {
			t.Errorf("%s %s = %d, want %d", tc.method, tc.path, w.Code, tc.want)
		}
	}
}
