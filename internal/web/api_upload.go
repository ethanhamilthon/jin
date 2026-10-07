package web

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"jin/internal/paths"
)

func (s *server) uploadRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/files/upload", api(func(r *http.Request) (any, error) {
		data, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, maxUpload))
		if err != nil {
			return nil, err
		}
		name := safeName(r.URL.Query().Get("name"))
		path, err := saveUpload(name, data)
		return map[string]string{"path": path, "name": name}, err
	}))
}

const maxUpload = 100 << 20

func safeName(name string) string {
	name = strings.TrimSpace(filepath.Base(strings.ReplaceAll(name, "\\", "/")))
	if name == "" || name == "." || name == "/" {
		return "file"
	}
	return name
}

// saveUpload keeps an attached file next to the pasted pictures.
func saveUpload(name string, data []byte) (string, error) {
	dir, err := paths.Global(".pasted")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, time.Now().Format("20060102-150405.000")+"-"+name)
	return path, os.WriteFile(path, data, 0o600)
}
