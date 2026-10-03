// Package sysprompt holds the three prompts that shape jin itself: the system
// prompt, the compaction prompt and the handoff prompt. Each has a built-in
// default. The user can change them in one file, ~/.jin/system-prompt.md,
// which has a section for each, introduced by a line "# system",
// "# compact" or "# handoff".
package sysprompt

import (
	_ "embed"
	"errors"
	"io/fs"
	"os"
	"strings"

	"jin/internal/paths"
)

//go:embed defaults/system.md
var defaultSystem string

//go:embed defaults/compact.md
var defaultCompact string

//go:embed defaults/handoff.md
var defaultHandoff string

// Names of the sections, in the order the file lists them.
const (
	SectionSystem  = "system"
	SectionCompact = "compact"
	SectionHandoff = "handoff"
)

// Sections are the texts in use. A section that the file leaves out or empty
// falls back to the built-in default.
type Sections struct {
	System, Compact, Handoff string
	// Custom is true when the file exists.
	Custom bool
}

// Path is the file the user edits.
func Path() (string, error) { return paths.Global("system-prompt.md") }

// Defaults are the built-in texts.
func Defaults() Sections {
	return Sections{
		System:  strings.TrimSpace(defaultSystem),
		Compact: strings.TrimSpace(defaultCompact),
		Handoff: strings.TrimSpace(defaultHandoff),
	}
}

// Load reads the file. A missing file means the defaults; an unreadable one
// is an error, and the defaults come back with it so the caller can go on.
func Load() (Sections, error) {
	sections := Defaults()
	path, err := Path()
	if err != nil {
		return sections, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return sections, nil
	}
	if err != nil {
		return sections, err
	}
	sections.Custom = true
	parsed := parse(string(data))
	for name, target := range map[string]*string{
		SectionSystem: &sections.System, SectionCompact: &sections.Compact, SectionHandoff: &sections.Handoff,
	} {
		if text := strings.TrimSpace(parsed[name]); text != "" {
			*target = text
		}
	}
	return sections, nil
}
