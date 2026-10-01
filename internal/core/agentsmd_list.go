package core

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"jin/internal/paths"
)

// AgentsFiles lists every AGENTS.md that applies to dir, empty ones included,
// in prompt order.
func AgentsFiles(dir string) []ContextFile {
	return existingFiles(dir)
}

// ProjectAgentsPath is where the project's own AGENTS.md lives.
func ProjectAgentsPath(dir string) string {
	return filepath.Join(dir, agentsFile)
}

// CreateProjectAgents makes an empty AGENTS.md in dir; an existing one is
// never touched.
func CreateProjectAgents(dir string) (string, error) {
	path := ProjectAgentsPath(dir)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return "", errors.New("AGENTS.md already exists here")
		}
		return "", err
	}
	return path, file.Close()
}

func existingFiles(dir string) []ContextFile {
	var files []ContextFile
	seen := map[string]bool{}
	for _, candidate := range candidates(dir) {
		data, err := os.ReadFile(candidate.Path)
		if err != nil || seen[candidate.Path] {
			continue
		}
		seen[candidate.Path] = true
		candidate.Content = strings.TrimSpace(string(data))
		files = append(files, candidate)
	}
	return files
}

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
	return append(list, ContextFile{Path: ProjectAgentsPath(dir), Kind: ContextProject})
}
