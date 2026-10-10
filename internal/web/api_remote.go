package web

import (
	"context"
	"net/http"
)

func (s *server) remoteRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/remote", api(func(r *http.Request) (any, error) {
		if s.remote == nil {
			return RemoteState{}, nil
		}
		return s.remote.state(), nil
	}))
	mux.HandleFunc("POST /api/remote", api(func(r *http.Request) (any, error) {
		var body struct{ Enabled bool }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		if s.remote == nil {
			return nil, nil
		}
		if err := s.remote.set(r.Context(), body.Enabled); err != nil {
			return nil, err
		}
		return s.remote.state(), nil
	}))
}

func (s *Service) SetRemote(ctx context.Context, enabled bool) error {
	return s.server.remote.set(ctx, enabled)
}
func (s *Service) Remote() RemoteState                { return s.server.remote.state() }
func (s *Service) Pair() (any, error)                 { return s.server.newPairing(nil) }
func (s *Service) Devices() (any, error)              { return s.server.db.Devices() }
func (s *Service) RenameDevice(id, name string) error { return s.server.db.RenameDevice(id, name) }
func (s *Service) RevokeDevice(id string) error {
	if err := s.server.db.RemoveDevice(id); err != nil {
		return err
	}
	s.server.hub.kick(id)
	return nil
}
