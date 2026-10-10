package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"jin/internal/session"
	"jin/internal/store"
	"jin/internal/web"
)

type webService struct {
	mu      sync.Mutex
	ctx     context.Context
	manager *session.Manager
	db      *store.DB
	version string
	service *web.Service
}

func (s *webService) route(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Version, Path string
		Options       web.Options
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body); err != nil {
		reply(w, 400, Result{Error: err.Error()})
		return
	}
	if body.Version != s.version {
		reply(w, 409, Result{Error: "client and daemon versions differ"})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.service == nil {
		dir, err := session.ProjectPath(body.Path, body.Path)
		if err != nil {
			reply(w, 400, Result{Error: err.Error()})
			return
		}
		if _, err := s.db.EnsureProject(dir); err != nil {
			reply(w, 500, Result{Error: err.Error()})
			return
		}
		s.service, err = web.StartService(s.ctx, s.manager, s.db, dir, s.version, body.Options)
		if err != nil {
			reply(w, 500, Result{Error: err.Error()})
			return
		}
	}
	if body.Options.Remote {
		if err := s.service.SetRemote(r.Context(), true); err != nil {
			reply(w, 400, Result{Error: err.Error()})
			return
		}
	}
	reply(w, 200, map[string]string{"url": s.service.URL()})
}

func (s *webService) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.service != nil {
		s.service.Close()
	}
}
