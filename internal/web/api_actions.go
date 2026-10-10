package web

import (
	"net/http"

	"jin/internal/session"
)

func (s *server) actionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/sessions/{id}/send", api(func(r *http.Request) (any, error) {
		var body struct {
			Text   string          `json:"text"`
			Images []session.Image `json:"images"`
			Files  []session.File  `json:"files"`
		}
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		return done(s.m.Send(r.PathValue("id"), body.Text, body.Images, body.Files))
	}))
	mux.HandleFunc("POST /api/sessions/{id}/shell", api(func(r *http.Request) (any, error) {
		var body struct{ Command string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		return done(s.m.Shell(r.PathValue("id"), body.Command))
	}))
	mux.HandleFunc("POST /api/sessions/{id}/answer", api(func(r *http.Request) (any, error) {
		var body struct{ Answers []string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		return done(s.m.Answer(r.PathValue("id"), body.Answers))
	}))
	mux.HandleFunc("POST /api/sessions/{id}/model", api(func(r *http.Request) (any, error) {
		var body struct{ Provider, Model, Effort string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		return done(s.m.SetModelProvider(r.PathValue("id"), body.Provider, body.Model, body.Effort))
	}))
	mux.HandleFunc("POST /api/sessions/{id}/fork", api(func(r *http.Request) (any, error) {
		var body struct{ Point int }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		return s.m.Fork(r.PathValue("id"), body.Point)
	}))
	simple := map[string]func(string) error{
		"stop": s.m.Interrupt, "compact": s.m.Compact, "handoff": s.m.Handoff, "reload": s.m.Reload,
		"seen": s.m.Seen, "focus": s.m.Focus,
	}
	for name, fn := range simple {
		mux.HandleFunc("POST /api/sessions/{id}/"+name, api(func(r *http.Request) (any, error) {
			return done(fn(r.PathValue("id")))
		}))
	}
	mux.HandleFunc("GET /api/sessions/{id}/points", api(func(r *http.Request) (any, error) {
		return s.m.Points(r.PathValue("id"))
	}))
	mux.HandleFunc("GET /api/sessions/{id}/context", api(func(r *http.Request) (any, error) {
		return s.m.Context(r.PathValue("id"))
	}))
}
