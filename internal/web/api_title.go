package web

import (
	"errors"
	"net/http"
)

func (s *server) titleRoutes(mux *http.ServeMux) {
	unavailable := api(func(r *http.Request) (any, error) {
		return nil, errors.New("title actions and settings are unavailable")
	})
	mux.HandleFunc("POST /api/sessions/{id}/generate-title", unavailable)
	mux.HandleFunc("POST /api/settings/title", unavailable)
}
