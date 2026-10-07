package web

import (
	"errors"
	"net/http"
	"os"
	"slices"

	"jin/internal/hooks"
	"jin/internal/prompts"
	"jin/internal/store"
)

// hookRef names a global hook, or with Project set a hook in .jin/hooks of
// the project Dir.
type hookRef struct {
	Name    string `json:"name"`
	Project bool   `json:"project"`
	Dir     string `json:"dir"`
	Content string `json:"content"`
	Trusted bool   `json:"trusted"`
}

func (h hookRef) path() (string, error) {
	if h.Project {
		return hooks.ProjectPath(h.Dir, h.Name)
	}
	return hooks.Path(h.Name)
}

func (h hookRef) key() string {
	if h.Project {
		return hooks.ProjectKey(h.Dir, h.Name)
	}
	return h.Name
}

type hookView struct {
	hookRef
	Enabled bool   `json:"enabled"`
	Preview string `json:"preview"`
}

func (s *server) hookRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/hooks", api(func(r *http.Request) (any, error) {
		dir := r.URL.Query().Get("dir")
		cfg, err := s.db.LoadConfig()
		if err != nil {
			return nil, err
		}
		trust, _ := s.db.HooksTrust(dir)
		global, err := hooks.List()
		project, _ := hooks.ListProject(dir)
		views := []hookView{}
		add := func(ref hookRef) {
			path, _ := ref.path()
			data, _ := os.ReadFile(path)
			on := !slices.Contains(cfg.HooksDisabled, ref.key()) && (!ref.Project || trust == store.Trusted)
			views = append(views, hookView{ref, on, prompts.Summary(string(data))})
		}
		for _, name := range global {
			add(hookRef{Name: name})
		}
		for _, name := range project {
			add(hookRef{Name: name, Project: true, Dir: dir})
		}
		return map[string]any{"hooks": views, "trust": trust}, err
	}))
	mux.HandleFunc("GET /api/hooks/text", api(func(r *http.Request) (any, error) {
		q := r.URL.Query()
		path, err := hookRef{Name: q.Get("name"), Project: q.Get("project") == "1", Dir: q.Get("dir")}.path()
		if err != nil {
			return nil, err
		}
		return readText(path)
	}))
	mux.HandleFunc("POST /api/hooks", s.hookChange(func(ref hookRef, _ store.Config) error {
		create := hooks.Create
		if ref.Project {
			create = func(name string) (string, error) { return hooks.CreateProject(ref.Dir, name) }
		}
		path, err := create(ref.Name)
		if err != nil {
			return err
		}
		return writeText(path, ref.Content)
	}))
	mux.HandleFunc("POST /api/hooks/delete", s.hookChange(func(ref hookRef, cfg store.Config) error {
		path, err := ref.path()
		if err != nil {
			return err
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		return s.db.SaveHooksDisabled(hooks.Forget(cfg.HooksDisabled, ref.key()))
	}))
	mux.HandleFunc("POST /api/hooks/toggle", s.hookChange(func(ref hookRef, cfg store.Config) error {
		if trust, _ := s.db.HooksTrust(ref.Dir); ref.Project && trust != store.Trusted {
			return errors.New("Trust the hooks of this project first")
		}
		return s.db.SaveHooksDisabled(hooks.Toggle(cfg.HooksDisabled, ref.key()))
	}))
	mux.HandleFunc("POST /api/hooks/trust", s.hookChange(func(ref hookRef, _ store.Config) error {
		return s.db.SaveHooksTrust(ref.Dir, ref.Trusted)
	}))
}

func (s *server) hookChange(fn func(ref hookRef, cfg store.Config) error) http.HandlerFunc {
	return s.changed(func(r *http.Request) error {
		var ref hookRef
		if err := decode(r, &ref); err != nil {
			return err
		}
		cfg, err := s.db.LoadConfig()
		if err != nil {
			return err
		}
		return fn(ref, cfg)
	})
}
