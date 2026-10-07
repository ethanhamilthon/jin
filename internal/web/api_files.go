package web

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"jin/internal/files"
	"jin/internal/paths"
)

const maxImage = 20 << 20

func (s *server) fileRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/files", api(func(r *http.Request) (any, error) {
		home, _ := os.UserHomeDir()
		dir := r.URL.Query().Get("dir")
		found := files.Candidates(r.URL.Query().Get("q"), home, dir, 30)
		out := []map[string]any{}
		for _, c := range found {
			out = append(out, map[string]any{"name": c.Name, "insert": files.Insert(c), "dir": c.IsDir})
		}
		return out, nil
	}))
	mux.HandleFunc("POST /api/images", api(func(r *http.Request) (any, error) {
		data, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, maxImage))
		if err != nil {
			return nil, err
		}
		path, err := saveImage(data)
		return map[string]string{"path": path}, err
	}))
}

// saveImage keeps a pasted or dropped picture where the TUI keeps its
// pasted images, so the model can read it.
func saveImage(data []byte) (string, error) {
	dir, err := paths.Global(".pasted")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "paste-"+time.Now().Format("20060102-150405.000")+imageExt(data))
	return path, os.WriteFile(path, data, 0o600)
}

func imageExt(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte("\xff\xd8\xff")):
		return ".jpg"
	case bytes.HasPrefix(data, []byte("GIF8")):
		return ".gif"
	case len(data) > 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return ".webp"
	case bytes.HasPrefix(data, []byte("BM")):
		return ".bmp"
	}
	return ".png"
}
