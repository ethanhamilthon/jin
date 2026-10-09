package core

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"jin/internal/paths"
)

// ContextFiles lists the non-empty AGENTS.md files that apply to dir, in
// prompt order: the global one, then parent directories from the outermost
// down, then the project's own. A file reached twice is listed once.
func ContextFiles(dir string) []ContextFile {
	var files []ContextFile
	seen := map[string]bool{}
	for _, candidate := range candidates(dir) {
		data, err := os.ReadFile(candidate.Path)
		if err != nil || seen[candidate.Path] {
			continue
		}
		seen[candidate.Path] = true
		candidate.Content = strings.TrimSpace(string(data))
		if candidate.Content != "" {
			files = append(files, candidate)
		}
	}
	return files
}

// candidates are the places an AGENTS.md may be, in prompt order.
func candidates(dir string) []ContextFile {
	var list []ContextFile
	if global, err := paths.Global(agentsFile); err == nil {
		list = append(list, ContextFile{Path: global, Kind: ContextGlobal})
	}
	var parents []string
	for current := dir; ; {
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		parents = append(parents, filepath.Join(parent, agentsFile))
		current = parent
	}
	slices.Reverse(parents)
	for _, path := range parents {
		list = append(list, ContextFile{Path: path, Kind: ContextParent})
	}
	return append(list, ContextFile{Path: filepath.Join(dir, agentsFile), Kind: ContextProject})
}
