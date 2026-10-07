package web

import (
	"net/http"
	"strconv"
)

func (s *server) entryImageRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/sessions/{id}/entries/{index}/image", func(w http.ResponseWriter, r *http.Request) {
		index, err := strconv.Atoi(r.PathValue("index"))
		if err != nil {
			http.Error(w, "bad index", http.StatusBadRequest)
			return
		}
		n, _ := strconv.Atoi(r.URL.Query().Get("n"))
		data, mime, err := s.m.EntryImage(r.PathValue("id"), index, n)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "private, max-age=3600")
		_, _ = w.Write(data)
	})
}

// localImageRoute serves a picture from disk for a Markdown image; only files
// whose content is an image are sent.
func (s *server) localImageRoute(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/local-image", func(w http.ResponseWriter, r *http.Request) {
		path := resolveLocal(r.URL.Query().Get("path"), r.URL.Query().Get("dir"))
		data, err := readImageFile(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", http.DetectContentType(data))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = w.Write(data)
	})
}
