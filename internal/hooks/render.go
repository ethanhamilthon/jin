package hooks

import (
	"os"
	"slices"
	"strings"
)

// Render joins the bodies of the enabled hooks, alphabetically by name, as
// plain text for the system prompt. Empty hooks add nothing.
func Render(disabled []string) string {
	names, _ := List()
	var bodies []string
	for _, name := range names {
		if slices.Contains(disabled, name) {
			continue
		}
		path, err := Path(name)
		if err != nil {
			continue
		}
		data, err := os.ReadFile(path)
		if body := strings.TrimSpace(string(data)); err == nil && body != "" {
			bodies = append(bodies, body)
		}
	}
	return strings.Join(bodies, "\n\n")
}

// Toggle switches one hook on or off in the list of disabled names.
func Toggle(disabled []string, name string) []string {
	if slices.Contains(disabled, name) {
		return Forget(disabled, name)
	}
	return append(slices.Clone(disabled), name)
}

// Forget drops a name from the disabled list, so a deleted hook that is
// created again starts out on.
func Forget(disabled []string, name string) []string {
	return slices.DeleteFunc(slices.Clone(disabled), func(n string) bool { return n == name })
}
