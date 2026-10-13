package web

import (
	"io/fs"
	"net/http"
)

// Dist is the built frontend for other shells, such as the desktop app.
func Dist() (fs.FS, bool) {
	root, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, false
	}
	_, err = fs.Stat(root, "index.html")
	return root, err == nil
}

// SPA serves root and falls back to index.html for page routes.
func SPA(root fs.FS) http.Handler {
	files := http.FileServerFS(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if name := r.URL.Path[1:]; name != "" {
			if _, err := fs.Stat(root, name); err != nil {
				r.URL.Path = "/"
			}
		}
		files.ServeHTTP(w, r)
	})
}
