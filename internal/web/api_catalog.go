package web

import (
	"jin/internal/sources"
	"net/http"
)

func (s *server) catalogRoute(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/models", api(func(r *http.Request) (any, error) {
		catalog, err := sources.List(r.Context(), s.db, r.URL.Query().Get("all") == "1")
		return s.catalogView(catalog), err
	}))
	mux.HandleFunc("POST /api/providers/{id}/enabled", api(func(r *http.Request) (any, error) {
		var body struct{ Enabled bool }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		if err := s.db.SetProviderEnabled(r.PathValue("id"), body.Enabled); err != nil {
			return nil, err
		}
		s.publish(map[string]string{"type": "config"})
		return done(s.m.ProvidersChanged())
	}))
}
