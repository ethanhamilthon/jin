package web

import (
	"jin/internal/sources"
	"net/http"
)

func (s *server) cliproxyAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/cliproxy/poll", api(func(r *http.Request) (any, error) {
		var body struct{ Profile, State string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		client, _, err := sources.Managed()
		if err != nil {
			return nil, err
		}
		status, err := client.Poll(r.Context(), body.Profile, body.State)
		if err == nil && status.Status == "ok" {
			s.publish(map[string]string{"type": "config"})
		}
		return status, err
	}))
	mux.HandleFunc("POST /api/cliproxy/cancel", api(func(r *http.Request) (any, error) {
		var body struct{ State string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		client, _, err := sources.Managed()
		if err != nil {
			return nil, err
		}
		return done(client.Cancel(r.Context(), body.State))
	}))
	mux.HandleFunc("POST /api/cliproxy/logout", s.changed(func(r *http.Request) error {
		var body struct{ Profile string }
		if err := decode(r, &body); err != nil {
			return err
		}
		client, _, err := sources.Managed()
		if err != nil {
			return err
		}
		return client.Logout(r.Context(), body.Profile)
	}))
}
