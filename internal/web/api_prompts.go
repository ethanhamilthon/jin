package web

import (
	"errors"
	"net/http"
	"slices"

	"jin/internal/prompts"
)

type promptView struct {
	Name    string `json:"name"`
	System  bool   `json:"system"`
	Preview string `json:"preview"`
	Enabled bool   `json:"enabled"`
}

type nameBody struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func (s *server) promptRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/prompts", api(func(r *http.Request) (any, error) {
		cfg, err := s.db.LoadConfig()
		if err != nil {
			return nil, err
		}
		infos, err := prompts.ListInfo()
		views := []promptView{}
		for _, info := range infos {
			views = append(views, promptView{info.Name, info.System, prompts.Preview(info.Name), !slices.Contains(cfg.PromptsDisabled, info.Name)})
		}
		return views, err
	}))
	mux.HandleFunc("GET /api/prompts/text", api(func(r *http.Request) (any, error) {
		path, err := userPrompt(r.URL.Query().Get("name"))
		if err != nil {
			return nil, err
		}
		return readText(path)
	}))
	mux.HandleFunc("POST /api/prompts", s.changed(func(r *http.Request) error {
		var body nameBody
		if err := decode(r, &body); err != nil {
			return err
		}
		path, err := prompts.Create(body.Name)
		if err != nil {
			return err
		}
		return writeText(path, body.Content)
	}))
	mux.HandleFunc("POST /api/prompts/delete", s.changed(func(r *http.Request) error {
		var body nameBody
		if err := decode(r, &body); err != nil {
			return err
		}
		if _, err := userPrompt(body.Name); err != nil {
			return err
		}
		if err := prompts.Delete(body.Name); err != nil {
			return err
		}
		cfg, err := s.db.LoadConfig()
		if err != nil {
			return err
		}
		return s.db.SavePromptsDisabled(slices.DeleteFunc(cfg.PromptsDisabled, func(n string) bool { return n == body.Name }))
	}))
	mux.HandleFunc("POST /api/prompts/toggle", s.changed(func(r *http.Request) error {
		var body nameBody
		if err := decode(r, &body); err != nil {
			return err
		}
		cfg, err := s.db.LoadConfig()
		if err != nil {
			return err
		}
		disabled := cfg.PromptsDisabled
		if slices.Contains(disabled, body.Name) {
			disabled = slices.DeleteFunc(disabled, func(n string) bool { return n == body.Name })
		} else {
			disabled = append(disabled, body.Name)
		}
		return s.db.SavePromptsDisabled(disabled)
	}))
}

func userPrompt(name string) (string, error) {
	if prompts.IsSystem(name) {
		return "", errors.New("system prompts cannot be changed")
	}
	return prompts.Path(name)
}
