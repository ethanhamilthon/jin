package web

import (
	"net/http"

	"jin/internal/store"
)

// titleRoutes names a session now, and saves the title settings.
func (s *server) titleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/sessions/{id}/generate-title", api(func(r *http.Request) (any, error) {
		title, err := s.m.GenerateTitle(r.Context(), r.PathValue("id"))
		if err != nil {
			return nil, err
		}
		return map[string]string{"title": title}, nil
	}))
	mux.HandleFunc("POST /api/settings/title", s.changed(func(r *http.Request) error {
		var body store.TitleSettings
		if err := decode(r, &body); err != nil {
			return err
		}
		return s.db.SaveTitle(body)
	}))
}
