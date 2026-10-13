package sysprompt

import "fmt"

// Reset retains its legacy signature without writing a file.
func Reset(texts map[string]string) (string, error) {
	for name := range texts {
		switch name {
		case SectionSystem, SectionCompact, SectionHandoff:
		default:
			return "", fmt.Errorf("unknown prompt %q", name)
		}
	}
	return Path()
}
