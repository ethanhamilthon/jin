package web

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"jin/internal/session"
)

type dirList struct {
	Path   string   `json:"path"`
	Parent string   `json:"parent"`
	Dirs   []string `json:"dirs"`
}

func (s *server) dirRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/dirs", api(func(r *http.Request) (any, error) {
		q := r.URL.Query()
		return listDirs(q.Get("path"), q.Get("hidden") == "1")
	}))
}

// listDirs lists the sub-folders of path, or of the home folder when path is
// empty. Folders starting with a dot are left out unless hidden is set.
func listDirs(path string, hidden bool) (dirList, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return dirList{}, err
	}
	if path == "" {
		path = home
	}
	dir, err := session.ProjectPath(path, home)
	if err != nil {
		return dirList{}, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return dirList{}, err
	}
	out := dirList{Path: dir, Dirs: []string{}}
	if parent := filepath.Dir(dir); parent != dir {
		out.Parent = parent
	}
	for _, entry := range entries {
		if !hidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if info, err := os.Stat(filepath.Join(dir, entry.Name())); err == nil && info.IsDir() {
			out.Dirs = append(out.Dirs, entry.Name())
		}
	}
	sort.Slice(out.Dirs, func(i, j int) bool {
		return strings.ToLower(out.Dirs[i]) < strings.ToLower(out.Dirs[j])
	})
	return out, nil
}
