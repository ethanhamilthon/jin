package daemon

import (
	"encoding/json"
	"net/http"
	"os"

	"jin/internal/web"
)

func (s *webService) remoteRoute(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Version, Action, Path, ID, Name string
		Enabled                         bool
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
		dir := body.Path
		if dir == "" {
			dir, _ = os.Getwd()
		}
		var err error
		s.service, err = web.StartService(s.ctx, s.manager, s.db, dir, s.version, web.Options{})
		if err != nil {
			reply(w, 500, Result{Error: err.Error()})
			return
		}
	}
	var value any
	var err error
	switch body.Action {
	case "state":
		value = s.service.Remote()
	case "set":
		err = s.service.SetRemote(r.Context(), body.Enabled)
		value = s.service.Remote()
	case "pair":
		value, err = s.service.Pair()
	case "devices":
		value, err = s.service.Devices()
	case "rename":
		err = s.service.RenameDevice(body.ID, body.Name)
	case "revoke":
		err = s.service.RevokeDevice(body.ID)
	default:
		reply(w, 400, Result{Error: "unknown remote action"})
		return
	}
	if err != nil {
		reply(w, 400, Result{Error: err.Error()})
		return
	}
	encoded, _ := json.Marshal(value)
	reply(w, 200, Result{Value: encoded})
}

func (s *webService) restore() {
	s.mu.Lock()
	defer s.mu.Unlock()
	enabled, _ := s.db.Setting("web.remote")
	if enabled != "true" {
		return
	}
	dir, _ := os.Getwd()
	var err error
	if s.service == nil {
		s.service, err = web.StartService(s.ctx, s.manager, s.db, dir, s.version, web.Options{})
	}
	if err == nil {
		_ = s.service.SetRemote(s.ctx, true)
	}
}
