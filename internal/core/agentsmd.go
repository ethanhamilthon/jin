package core

import "strings"

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
