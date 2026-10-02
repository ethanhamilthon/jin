package files

import (
	"os"
	"path/filepath"
	"strings"
)

type Candidate struct {
	Name  string
	Path  string
	IsDir bool
}

func pathFor(raw, home, cwd string) string {
	p := raw
	if strings.HasPrefix(p, "~/") {
		p = filepath.Join(home, p[2:])
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(cwd, p)
	}
	return filepath.Clean(p)
}

func Candidates(raw, home, cwd string, limit int) []Candidate {
	if limit <= 0 {
		return nil
	}
	// The typed text splits at its last "/": the part before it is the
	// directory to list and is kept exactly as typed in every result.
	rawDir, prefix := "", raw
	if i := strings.LastIndexByte(raw, '/'); i >= 0 {
		rawDir, prefix = raw[:i+1], raw[i+1:]
	}
	dir := cwd
	if rawDir != "" {
		dir = pathFor(rawDir, home, cwd)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var dirs, files []Candidate
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			continue
		}
		isDir := e.IsDir()
		if !isDir && e.Type()&os.ModeSymlink != 0 {
			if st, err := os.Stat(filepath.Join(dir, name)); err == nil && st.IsDir() {
				isDir = true
			}
		}
		insertPath := rawDir + name
		if isDir {
			insertPath += "/"
		}
		c := Candidate{Name: name, Path: insertPath, IsDir: isDir}
		if isDir {
			dirs = append(dirs, c)
		} else {
			files = append(files, c)
		}
	}
	all := append(dirs, files...)
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}

func Insert(c Candidate) string {
	if strings.ContainsAny(c.Path, " \t\n\r\"") {
		return `"` + strings.ReplaceAll(c.Path, `"`, `\"`) + `"`
	}
	return c.Path
}
