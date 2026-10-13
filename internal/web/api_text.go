package web

import (
	"errors"
	"net/http"
	"os"

	"jin/internal/sysprompt"
)

func (s *server) sysPromptRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/sysprompt", api(func(r *http.Request) (any, error) {
		sections := sysprompt.Defaults()
		return textView{Content: sections.System + "\n\n# compact\n\n" + sections.Compact + "\n\n# handoff\n\n" + sections.Handoff}, nil
	}))
	unavailable := api(func(r *http.Request) (any, error) {
		return nil, errors.New("system prompt editing is unavailable")
	})
	mux.HandleFunc("POST /api/sysprompt", unavailable)
	mux.HandleFunc("POST /api/sysprompt/reset", unavailable)
}

type textView struct {
	Path    string `json:"path,omitempty"`
	Content string `json:"content"`
}

func readText(path string) (textView, error) {
	data, err := os.ReadFile(path)
	return textView{Path: path, Content: string(data)}, err
}

func writeText(path, content string) error { return os.WriteFile(path, []byte(content), 0o600) }
