package core

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"jin/internal/paths"
)

type ContextKind string

const (
	ContextGlobal  ContextKind = "global"
	ContextParent  ContextKind = "parent"
	ContextProject ContextKind = "project"
)

const agentsFile = "AGENTS.md"

// ContextFile is one AGENTS.md that goes into the system prompt.
type ContextFile struct {
	Path    string
	Kind    ContextKind
	Content string
}

func (f ContextFile) source() string {
	switch f.Kind {
	case ContextGlobal:
		return "global, applies to every project"
	case ContextParent:
		return "parent directory, not the current project"
	default:
		return "current project"
	}
}

// ContextFiles lists the non-empty AGENTS.md files that apply to dir, in
// prompt order: the global one, then parent directories from the outermost
// down, then the project's own.
func ContextFiles(dir string) []ContextFile {
	var files []ContextFile
	seen := map[string]bool{}
	add := func(path string, kind ContextKind) {
		data, err := os.ReadFile(path)
		content := strings.TrimSpace(string(data))
		if err != nil || content == "" || seen[path] {
			return
		}
		seen[path] = true
		files = append(files, ContextFile{Path: path, Kind: kind, Content: content})
	}
	if global, err := paths.Global(agentsFile); err == nil {
		add(global, ContextGlobal)
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
		add(path, ContextParent)
	}
	add(filepath.Join(dir, agentsFile), ContextProject)
	return files
}

func renderContext(files []ContextFile) string {
	if len(files) == 0 {
		return "(none found)"
	}
	blocks := make([]string, len(files))
	for i, f := range files {
		blocks[i] = "<agents-md path=\"" + f.Path + "\" source=\"" + f.source() + "\">\n" + f.Content + "\n</agents-md>"
	}
	return strings.Join(blocks, "\n\n")
}
