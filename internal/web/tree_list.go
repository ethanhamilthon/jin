package web

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const maxEntries = 3000

type treeEntry struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir"`
	Size int64  `json:"size"`
}

func (s *server) treeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/tree", api(func(r *http.Request) (any, error) {
		q := r.URL.Query()
		return listTree(r.Context(), q.Get("dir"), q.Get("path"), q.Get("hidden") == "1")
	}))
	s.fileViewRoutes(mux)
}

// listTree lists one folder of a project, folders first. Unless hidden is
// set, .git and what the project's gitignore rules ignore are left out.
func listTree(ctx context.Context, dir, rel string, hidden bool) ([]treeEntry, error) {
	full, err := inProject(dir, rel)
	if err != nil {
		return nil, err
	}
	items, err := os.ReadDir(full)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, item := range items {
		if hidden || item.Name() != ".git" {
			names = append(names, item.Name())
		}
	}
	if !hidden {
		ignored := gitIgnored(ctx, dir, rel, names)
		names = filterNames(names, ignored)
	}
	out := make([]treeEntry, 0, len(names))
	for _, name := range names {
		info, err := os.Stat(filepath.Join(full, name))
		if err != nil {
			continue
		}
		out = append(out, treeEntry{Name: name, Dir: info.IsDir(), Size: info.Size()})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Dir != out[j].Dir {
			return out[i].Dir
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	if len(out) > maxEntries {
		out = out[:maxEntries]
	}
	return out, nil
}

func filterNames(names []string, ignored map[string]bool) []string {
	var out []string
	for _, name := range names {
		if !ignored[name] {
			out = append(out, name)
		}
	}
	return out
}

// gitIgnored asks git which of the names in folder rel it ignores. Outside a
// git repository, or when git is missing, nothing is ignored.
func gitIgnored(ctx context.Context, dir, rel string, names []string) map[string]bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var input bytes.Buffer
	for _, name := range names {
		input.WriteString(filepath.ToSlash(filepath.Join(rel, name)) + "\x00")
	}
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "check-ignore", "-z", "--stdin")
	cmd.Stdin = &input
	out, _ := cmd.Output()
	ignored := map[string]bool{}
	for _, path := range strings.Split(string(out), "\x00") {
		if path != "" {
			ignored[filepath.Base(path)] = true
		}
	}
	return ignored
}
