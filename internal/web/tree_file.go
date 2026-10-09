package web

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const maxPreview = 1 << 20

type fileView struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Kind    string `json:"kind"` // text, markdown, image, binary or large
	Content string `json:"content,omitempty"`
}

func (s *server) fileViewRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/tree/file", api(func(r *http.Request) (any, error) {
		q := r.URL.Query()
		return viewFile(q.Get("dir"), q.Get("path"))
	}))
	mux.HandleFunc("GET /api/tree/raw", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		full, err := inProject(q.Get("dir"), q.Get("path"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data, err := readImageFile(full)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", http.DetectContentType(data))
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(data)
	})
}

// viewFile reads a file of a project for the preview: its text up to 1 MB,
// or only what kind it is (a picture, a binary file, a file that is too big).
func viewFile(dir, rel string) (fileView, error) {
	full, err := inProject(dir, rel)
	if err != nil {
		return fileView{}, err
	}
	info, err := os.Stat(full)
	if err != nil {
		return fileView{}, err
	}
	if !info.Mode().IsRegular() {
		return fileView{}, errors.New("not a regular file")
	}
	view := fileView{Name: filepath.Base(full), Path: filepath.ToSlash(rel), Size: info.Size()}
	if info.Size() > maxPreview {
		view.Kind = "large"
		if isPicture(full) {
			view.Kind = "image"
		}
		return view, nil
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return fileView{}, err
	}
	switch {
	case strings.HasPrefix(http.DetectContentType(data), "image/"):
		view.Kind = "image"
	case !looksLikeText(data):
		view.Kind = "binary"
	case strings.HasSuffix(strings.ToLower(full), ".md") || strings.HasSuffix(strings.ToLower(full), ".markdown"):
		view.Kind, view.Content = "markdown", string(data)
	default:
		view.Kind, view.Content = "text", string(data)
	}
	return view, nil
}

func isPicture(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp":
		return true
	}
	return false
}

func looksLikeText(data []byte) bool {
	head := data[:min(len(data), 8192)]
	return !bytes.Contains(head, []byte{0}) && utf8.Valid(data)
}
