package web

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"time"

	"jin/internal/export"
	"jin/internal/session"
	"jin/internal/store"
)

type sessionRow struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	Path      string      `json:"path"`
	Model     string      `json:"model"`
	UpdatedAt time.Time   `json:"updated_at"`
	Usage     store.Usage `json:"usage"`
	Unread    bool        `json:"unread"`
	Snippet   string      `json:"snippet,omitempty"`
}

func (s *server) sessionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/sessions", api(s.listSessions))
	mux.HandleFunc("POST /api/sessions", api(func(r *http.Request) (any, error) {
		var body struct{ Path string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		dir, err := session.ProjectPath(body.Path, s.dir)
		if err != nil {
			return nil, err
		}
		if _, err := s.db.EnsureProject(dir); err != nil {
			return nil, err
		}
		return s.m.Create(dir)
	}))
	mux.HandleFunc("POST /api/sessions/{id}/open", api(func(r *http.Request) (any, error) {
		return s.m.Open(r.PathValue("id"))
	}))
	mux.HandleFunc("GET /api/sessions/{id}", api(func(r *http.Request) (any, error) {
		return s.m.Snapshot(r.PathValue("id"))
	}))
	mux.HandleFunc("POST /api/sessions/{id}/close", api(func(r *http.Request) (any, error) {
		return map[string]bool{"closed": s.m.Close(r.PathValue("id"))}, nil
	}))
	mux.HandleFunc("GET /api/sessions/{id}/export", api(func(r *http.Request) (any, error) {
		var out, errOut bytes.Buffer
		if export.Main([]string{r.PathValue("id"), "--md"}, s.db, &out, &errOut) != 0 {
			return nil, errors.New(strings.TrimSpace(errOut.String()))
		}
		return map[string]string{"markdown": out.String()}, nil
	}))
}

// listSessions lists the sessions of a project (?path=, default the start
// directory) or of all projects (?all=1), optionally matching ?q=.
func (s *server) listSessions(r *http.Request) (any, error) {
	path, all, query := r.URL.Query().Get("path"), r.URL.Query().Get("all") == "1", r.URL.Query().Get("q")
	if path == "" {
		path = s.dir
	}
	var results []store.SessionResult
	var err error
	if words := strings.Fields(query); len(words) > 0 {
		results, err = s.db.SearchSessions(path, all, words)
	} else {
		results, err = s.db.ListSessions(path, all)
	}
	if err != nil {
		return nil, err
	}
	unread, _ := s.db.AllUnread()
	rows := []sessionRow{}
	for _, res := range results {
		row := sessionRow{ID: res.ID, Title: res.Title, Path: res.Path, UpdatedAt: res.UpdatedAt, Snippet: res.Snippet, Unread: unread[res.ID]}
		if rec, ok, _ := s.db.GetSession(res.ID); ok {
			row.Model, row.Usage = rec.Model, rec.Usage
		}
		rows = append(rows, row)
	}
	return rows, nil
}
