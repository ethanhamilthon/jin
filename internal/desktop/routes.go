//go:build desktop

package desktop

import (
	"net/http"
	"path"
	"strings"
)

var routes = []string{
	"GET /api/state", "GET /api/events", "GET /api/data", "POST /api/data",
	"GET /api/devices", "POST /api/devices/pair", "PATCH /api/devices/{id}", "DELETE /api/devices/{id}",
	"GET /api/remote", "POST /api/remote",
	"GET /api/dirs", "GET /api/files", "POST /api/files/upload", "POST /api/images", "GET /api/local-image",
	"GET /api/models", "GET /api/projects", "POST /api/projects", "POST /api/projects/archive", "POST /api/projects/rename",
	"GET /api/hooks", "GET /api/hooks/text", "POST /api/hooks", "POST /api/hooks/delete", "POST /api/hooks/toggle", "POST /api/hooks/trust",
	"GET /api/prompts", "GET /api/prompts/text", "POST /api/prompts", "POST /api/prompts/delete", "POST /api/prompts/toggle",
	"POST /api/providers", "GET /api/providers/name", "POST /api/providers/probe", "POST /api/providers/probe-efforts",
	"GET /api/providers/{id}/models", "GET /api/providers/{id}/efforts",
	"POST /api/providers/{id}/activate", "POST /api/providers/{id}/delete", "POST /api/providers/{id}/enabled",
	"GET /api/sessions", "POST /api/sessions", "GET /api/sessions/{id}", "GET /api/sessions/{id}/context",
	"GET /api/sessions/{id}/points", "GET /api/sessions/{id}/export", "GET /api/sessions/{id}/entries/{index}/image",
	"POST /api/settings/accent", "POST /api/settings/fold", "POST /api/settings/scope", "GET /api/settings/scope",
	"POST /api/settings/sound", "POST /api/settings/tools",
	"GET /api/sounds", "GET /api/sounds/{file}",
	"GET /api/tasks", "GET /api/tasks/{id}/output", "POST /api/tasks/{id}/stop",
	"GET /api/tree", "GET /api/tree/file", "GET /api/tree/raw",
}

var sessionActions = []string{"send", "shell", "answer", "model", "fork", "handoff", "stop", "compact", "reload", "seen", "focus", "open", "close"}

func newRouter(api http.Handler) http.Handler {
	mux := http.NewServeMux()
	for _, route := range routes {
		mux.Handle(route, api)
	}
	for _, action := range sessionActions {
		mux.Handle("POST /api/sessions/{id}/"+action, api)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path.Clean(r.URL.Path) || r.URL.RawPath != "" || strings.Contains(r.URL.Path, "//") {
			http.NotFound(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
