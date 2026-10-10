package daemon

import (
	"encoding/json"
	"net/http"
	"os"

	"jin/internal/session"
	"jin/internal/tasks"
)

func routes(version string, manager *session.Manager, stop func()) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("version") != version {
			reply(w, http.StatusConflict, map[string]string{"error": "client and daemon versions differ"})
			return
		}
		eventStream(manager)(w, r)
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		reply(w, http.StatusOK, Status{version, os.Getpid(), manager.Working(), len(tasks.Shared().Running(""))})
	})
	mux.HandleFunc("POST /stop", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("version") != version {
			reply(w, http.StatusConflict, map[string]string{"error": "client and daemon versions differ"})
			return
		}
		if r.URL.Query().Get("force") != "true" && (manager.Working() || len(tasks.Shared().Running("")) > 0) {
			reply(w, http.StatusConflict, map[string]string{"error": "daemon is busy; use jin daemon stop --force to interrupt agents and tasks"})
			return
		}
		reply(w, http.StatusOK, map[string]bool{"stopping": true})
		stop()
	})
	return mux
}

func reply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
