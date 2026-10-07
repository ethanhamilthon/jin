package web

import (
	"context"
	"net/http"
	"sync"

	"jin/internal/datadir"
	"jin/internal/session"
	"jin/internal/store"
)

// server is one `jin web`: the live sessions, the open pages and the
// directory it started in.
type server struct {
	ctx     context.Context
	db      *store.DB
	m       *session.Manager
	hub     *hub
	dir     string
	version string
	quit    func()

	mu     sync.Mutex
	action *datadir.Action
	latest string
}

func (s *server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/events", s.hub.serve)
	s.stateRoutes(mux)
	s.sessionRoutes(mux)
	s.actionRoutes(mux)
	s.projectRoutes(mux)
	s.providerRoutes(mux)
	s.settingRoutes(mux)
	s.promptRoutes(mux)
	s.sysPromptRoutes(mux)
	s.hookRoutes(mux)
	s.taskRoutes(mux)
	s.fileRoutes(mux)
	s.uploadRoutes(mux)
	s.entryImageRoutes(mux)
	s.localImageRoute(mux)
	s.dataRoutes(mux)
	if files, ok := assets(); ok {
		mux.Handle("GET /", files)
	} else {
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, noUI, http.StatusNotFound)
		})
	}
	return mux
}

const noUI = "This build has no web UI: run make build or install a release."

func (s *server) publish(ev any) { s.hub.publish(ev) }
