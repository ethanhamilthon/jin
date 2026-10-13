package web

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"jin/internal/sounds"
)

func (s *server) soundRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/sounds", api(func(r *http.Request) (any, error) {
		return sounds.Catalog(), nil
	}))
	mux.HandleFunc("GET /api/sounds/{file}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := strings.CutSuffix(r.PathValue("file"), ".wav")
		if !ok || sounds.Normalize(id) != id {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("Cache-Control", "private, max-age=86400")
		http.ServeContent(w, r, id+".wav", time.Time{}, bytes.NewReader(sounds.WAV(id)))
	})
}
