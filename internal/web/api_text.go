package web

import (
	"net/http"
	"os"

	"jin/internal/sysprompt"
)

func (s *server) sysPromptRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/sysprompt", api(func(r *http.Request) (any, error) {
		path, err := sysprompt.Create()
		if err != nil {
			return nil, err
		}
		return readText(path)
	}))
	mux.HandleFunc("POST /api/sysprompt", s.changed(func(r *http.Request) error {
		var body nameBody
		if err := decode(r, &body); err != nil {
			return err
		}
		path, err := sysprompt.Create()
		if err != nil {
			return err
		}
		return writeText(path, body.Content)
	}))
}

type textView struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func readText(path string) (textView, error) {
	data, err := os.ReadFile(path)
	return textView{path, string(data)}, err
}

func writeText(path, content string) error { return os.WriteFile(path, []byte(content), 0o600) }
