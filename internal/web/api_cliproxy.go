package web

import (
	"errors"
	"jin/internal/cliproxy"
	"jin/internal/sources"
	"net"
	"net/http"
	"strings"
)

func localRequest(r *http.Request) bool {
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return host == "localhost" || net.ParseIP(strings.Trim(host, "[]")).IsLoopback()
}

func (s *server) cliproxyRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/cliproxy", api(func(r *http.Request) (any, error) {
		_, root, err := sources.Managed()
		return map[string]any{"installation": cliproxy.Installed(root), "local_login": localRequest(r)}, err
	}))
	mux.HandleFunc("GET /api/cliproxy/versions", api(func(r *http.Request) (any, error) { return cliproxy.Versions(r.Context()) }))
	mux.HandleFunc("POST /api/cliproxy/install", s.changed(func(r *http.Request) error {
		var body struct{ Version string }
		if err := decode(r, &body); err != nil {
			return err
		}
		if body.Version == "" {
			body.Version = cliproxy.DefaultVersion
		}
		_, root, err := sources.Managed()
		if err != nil {
			return err
		}
		return cliproxy.InstallInitial(r.Context(), root, body.Version)
	}))
	mux.HandleFunc("POST /api/cliproxy/update", s.changed(func(r *http.Request) error {
		var body struct{ Version string }
		if err := decode(r, &body); err != nil {
			return err
		}
		client, _, err := sources.Managed()
		if err != nil {
			return err
		}
		return client.Update(r.Context(), body.Version)
	}))
	mux.HandleFunc("GET /api/cliproxy/accounts", api(func(r *http.Request) (any, error) {
		client, root, err := sources.Managed()
		if err != nil {
			return nil, err
		}
		if cliproxy.Installed(root).Version == "" {
			return []cliproxy.Account{}, nil
		}
		return client.Accounts(r.Context())
	}))
	mux.HandleFunc("POST /api/cliproxy/providers", s.changed(func(r *http.Request) error {
		var body struct{ Profile string }
		if err := decode(r, &body); err != nil {
			return err
		}
		if err := sources.AddManaged(s.db, body.Profile); err != nil {
			return err
		}
		return s.m.ProvidersChanged()
	}))
	mux.HandleFunc("POST /api/cliproxy/login", api(func(r *http.Request) (any, error) {
		if !localRequest(r) {
			return nil, errors.New("sign in on the computer running Jin; phones can manage connected providers")
		}
		var body struct{ Profile string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		client, _, err := sources.Managed()
		if err != nil {
			return nil, err
		}
		return client.Login(r.Context(), body.Profile)
	}))
	s.cliproxyAuthRoutes(mux)
}
