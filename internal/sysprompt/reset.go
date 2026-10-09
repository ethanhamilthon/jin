package sysprompt

import (
	"fmt"
	"os"
)

// Reset replaces the named sections of the user's file with the given texts
// and keeps the others as they are. It creates the file when it is missing,
// and returns its path.
func Reset(texts map[string]string) (string, error) {
	path, err := Create()
	if err != nil {
		return "", err
	}
	current, err := Load()
	if err != nil {
		return "", err
	}
	for name, text := range texts {
		switch name {
		case SectionSystem:
			current.System = text
		case SectionCompact:
			current.Compact = text
		case SectionHandoff:
			current.Handoff = text
		default:
			return "", fmt.Errorf("unknown prompt %q", name)
		}
	}
	return path, os.WriteFile(path, []byte(current.Render()), 0o600)
}
