package core

import (
	"os"
	"path/filepath"
	"strings"
)

// ContextFiles lists the AGENTS.md of dir when it has text. Parent folders and
// the global file are not read.
func ContextFiles(dir string) []ContextFile {
	path := filepath.Join(dir, agentsFile)
	data, err := os.ReadFile(path)
	if content := strings.TrimSpace(string(data)); err == nil && content != "" {
		return []ContextFile{{Path: path, Kind: ContextProject, Content: content}}
	}
	return nil
}
