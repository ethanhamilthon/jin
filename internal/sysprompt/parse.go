package sysprompt

import (
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

// Create retains its legacy signature without creating or changing a file.
func Create() (string, error) { return Path() }
