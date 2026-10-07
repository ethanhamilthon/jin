package web

import (
	"errors"
	"net/http"
	"regexp"
	"slices"

	"jin/internal/session"
	"jin/internal/store"
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func (s *server) settingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/settings/sound", s.changed(func(r *http.Request) error {
		var body soundView
		if err := decode(r, &body); err != nil {
			return err
		}
		return s.db.SaveSound(store.Sound{Enabled: body.Enabled, OnlyBlur: body.OnlyBlur, Volume: min(100, max(0, body.Volume))})
	}))
	mux.HandleFunc("POST /api/settings/accent", s.changed(func(r *http.Request) error {
		var body struct{ Accent string }
		if err := decode(r, &body); err != nil {
			return err
		}
		if body.Accent != "" && !hexColor.MatchString(body.Accent) {
			return errors.New("accent must be a color like #52a8ff")
		}
		return s.db.SetSetting(accentKey, body.Accent)
	}))
	mux.HandleFunc("POST /api/settings/tools", s.changed(func(r *http.Request) error {
		var body struct {
			Name    string
			Enabled bool
		}
		if err := decode(r, &body); err != nil {
			return err
		}
		cfg, err := s.db.LoadConfig()
		if err != nil {
			return err
		}
		disabled := slices.DeleteFunc(slices.Clone(cfg.ToolsDisabled), func(n string) bool { return n == body.Name })
		if !body.Enabled {
			disabled = append(disabled, body.Name)
		}
		return s.db.SaveToolsDisabled(disabled)
	}))
	mux.HandleFunc("POST /api/settings/scope", s.changed(func(r *http.Request) error {
		var body struct{ Provider, Model string }
		if err := decode(r, &body); err != nil {
			return err
		}
		client, scope, err := s.clientOf(body.Provider)
		if err != nil {
			return err
		}
		all, err := client.Models(r.Context())
		if err != nil {
			return err
		}
		return s.db.SaveScopeFor(body.Provider, session.ToggleScope(all, scope, body.Model))
	}))
	mux.HandleFunc("GET /api/settings/scope", api(func(r *http.Request) (any, error) {
		scope, err := s.db.LoadScopeFor(r.URL.Query().Get("provider"))
		return append([]string{}, scope...), err
	}))
}

// changed runs a settings change and tells the pages to load the
// configuration again.
func (s *server) changed(fn func(r *http.Request) error) http.HandlerFunc {
	return api(func(r *http.Request) (any, error) {
		if err := fn(r); err != nil {
			return nil, err
		}
		s.publish(map[string]string{"type": "config"})
		return nil, nil
	})
}
