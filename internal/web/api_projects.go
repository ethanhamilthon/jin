package web

import (
	"errors"
	"net/http"
	"path/filepath"

	"jin/internal/session"
	"jin/internal/store"
)

type projectView struct {
	store.Project
	Sessions int  `json:"sessions"`
	Unread   bool `json:"unread"`
}

func (s *server) projectRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/projects", api(func(r *http.Request) (any, error) {
		projects, err := s.db.Projects()
		if err != nil {
			return nil, err
		}
		owners, _ := s.db.SessionProjects()
		unread, _ := s.db.AllUnread()
		views := []projectView{}
		for _, p := range projects {
			view := projectView{Project: p}
			if view.Name == "" {
				view.Name = filepath.Base(p.Path)
			}
			for id, path := range owners {
				if path == p.Path {
					view.Sessions++
					view.Unread = view.Unread || unread[id]
				}
			}
			views = append(views, view)
		}
		return views, nil
	}))
	mux.HandleFunc("POST /api/projects", api(func(r *http.Request) (any, error) {
		var body struct{ Path string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		dir, err := session.ProjectPath(body.Path, s.dir)
		if err != nil {
			return nil, err
		}
		project, err := s.db.EnsureProject(dir)
		if err == nil {
			err = s.db.SetProjectArchived(dir, false)
		}
		s.publish(map[string]string{"type": "projects"})
		return project, err
	}))
	mux.HandleFunc("POST /api/projects/remove", api(func(r *http.Request) (any, error) {
		var body struct{ Path string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		if body.Path == s.dir {
			return nil, errors.New("Cannot remove the project jin web started in")
		}
		err := s.db.RemoveProject(body.Path)
		s.publish(map[string]string{"type": "projects"})
		return done(err)
	}))
	mux.HandleFunc("POST /api/projects/archive", api(func(r *http.Request) (any, error) {
		var body struct {
			Path     string
			Archived bool
		}
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		err := s.db.SetProjectArchived(body.Path, body.Archived)
		s.publish(map[string]string{"type": "projects"})
		return done(err)
	}))
	mux.HandleFunc("POST /api/projects/rename", api(func(r *http.Request) (any, error) {
		var body struct{ Path, Name string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		err := s.db.RenameProject(body.Path, body.Name)
		s.publish(map[string]string{"type": "projects"})
		return done(err)
	}))
}
