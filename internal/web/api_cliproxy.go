package web

import (
	"encoding/json"
	"net/http"
)

func (s *server) cliproxyRoutes(mux *http.ServeMux) {
	unavailable := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "CLIProxyAPI support was removed"})
	}
	mux.HandleFunc("GET /api/cliproxy", unavailable)
	mux.HandleFunc("GET /api/cliproxy/versions", unavailable)
	mux.HandleFunc("POST /api/cliproxy/install", unavailable)
	mux.HandleFunc("POST /api/cliproxy/update", unavailable)
	mux.HandleFunc("GET /api/cliproxy/accounts", unavailable)
	mux.HandleFunc("POST /api/cliproxy/providers", unavailable)
	mux.HandleFunc("POST /api/cliproxy/login", unavailable)
	mux.HandleFunc("POST /api/cliproxy/poll", unavailable)
	mux.HandleFunc("POST /api/cliproxy/cancel", unavailable)
	mux.HandleFunc("POST /api/cliproxy/logout", unavailable)
}
