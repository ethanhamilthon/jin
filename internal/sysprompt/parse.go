package sysprompt

import (
	_ "embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// parse splits the file at its section lines. Text before the first section
// line is ignored. A section line is exactly "# system", "# compact" or
// "# handoff"; every other line, including other headings, belongs to the
// section above it. A name that appears twice keeps its last text.
func parse(text string) map[string]string {
	out := map[string]string{}
	var current string
	var body []string
	flush := func() {
		if current != "" {
			out[current] = strings.Join(body, "\n")
		}
		body = nil
	}
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		switch strings.TrimRight(line, " \t") {
		case "# " + SectionSystem:
			flush()
			current = SectionSystem
		case "# " + SectionCompact:
			flush()
			current = SectionCompact
		case "# " + SectionHandoff:
			flush()
			current = SectionHandoff
		default:
			body = append(body, line)
		}
	}
	flush()
	return out
}

// Render writes the three sections in file form.
func (s Sections) Render() string {
	return "# " + SectionSystem + "\n\n" + strings.TrimSpace(s.System) + "\n\n" +
		"# " + SectionCompact + "\n\n" + strings.TrimSpace(s.Compact) + "\n\n" +
		"# " + SectionHandoff + "\n\n" + strings.TrimSpace(s.Handoff) + "\n"
}

// Create writes the file from the current defaults, unless it exists, and
// returns its path.
func Create() (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return path, nil
	}
	if err != nil {
		return "", err
	}
	_, werr := file.WriteString(Defaults().Render())
	if cerr := file.Close(); werr == nil {
		werr = cerr
	}
	return path, werr
}
