package hooks

import (
	"os"
	"slices"
	"strings"
)

// Hook is a switched-on hook with something to say. Body is the raw text of
// the file; commands in it are run when a session starts (see package dyn).
type Hook struct {
	Name, Body string
	// Project marks a hook from the repository's .jin/hooks folder.
	Project bool
}

// Active lists the hooks that go into a system prompt, alphabetically: the
// ones not disabled and not empty.
func Active(disabled []string) []Hook {
	var active []Hook
	names, _ := List()
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
			active = append(active, Hook{Name: name, Body: body})
		}
	}
	return active
}

// Render joins the bodies of the active hooks as plain text for the system prompt.
func Render(disabled []string) string {
	var bodies []string
	for _, hook := range Active(disabled) {
		bodies = append(bodies, hook.Body)
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
