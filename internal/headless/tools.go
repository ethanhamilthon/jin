package headless

import (
	"fmt"
	"slices"

	"jin/internal/tools"
)

// toolNames is the tool set of a headless run: the enabled tools, never
// ask_user (nobody can answer), narrowed by the flags. A flag can only remove
// tools.
func toolNames(disabled []string, opt Options) ([]string, error) {
	known := tools.Catalog()
	for _, name := range append(slices.Clone(opt.Tools), opt.Exclude...) {
		if !slices.Contains(known, name) {
			return nil, fmt.Errorf("unknown tool %q (known: %v)", name, known)
		}
	}
	var out []string
	for _, name := range tools.Without(disabled) {
		switch {
		case name == "ask_user", opt.NoTools:
		case opt.ToolsSet && !slices.Contains(opt.Tools, name):
		case slices.Contains(opt.Exclude, name):
		default:
			out = append(out, name)
		}
	}
	return out, nil
}
