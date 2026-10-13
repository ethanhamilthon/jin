// Package sysprompt holds Jin's built-in system, compaction and handoff prompts.
package sysprompt

import (
	_ "embed"
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

// Sections contains the built-in prompt texts.
type Sections struct {
	System, Compact, Handoff string
	Custom                   bool
}

// Path retains the old file location API without reading or creating the file.
func Path() (string, error) { return paths.Global("system-prompt.md") }

// Defaults are the built-in texts.
func Defaults() Sections {
	return Sections{
		System:  strings.TrimSpace(defaultSystem),
		Compact: strings.TrimSpace(defaultCompact),
		Handoff: strings.TrimSpace(defaultHandoff),
	}
}

// Load preserves its legacy signature and returns only built-in prompts.
func Load() (Sections, error) { return Defaults(), nil }
