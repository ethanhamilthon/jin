package hooks

import "path/filepath"

// ActiveIn lists the hooks of a session in dir: the global ones, then the
// project ones when trusted is set. Disabled and empty hooks are left out.
func ActiveIn(dir string, disabled []string, trusted bool) []Hook {
	active, _ := LoadIn(dir, disabled, trusted)
	return active
}

// LoadIn is ActiveIn plus a warning for each hook folder or file that exists
// but cannot be read.
func LoadIn(dir string, disabled []string, trusted bool) ([]Hook, []string) {
	active, warnings := load(disabled)
	if !trusted {
		return active, warnings
	}
	names, err := ListProject(dir)
	warnings = append(warnings, listWarning(err)...)
	for _, name := range names {
		if contains(disabled, ProjectKey(dir, name)) {
			continue
		}
		body, warning := readBody(filepath.Join(ProjectDir(dir), name+ext))
		if warning != "" {
			warnings = append(warnings, warning)
		}
		if body != "" {
			active = append(active, Hook{Name: name, Body: body, Project: true})
		}
	}
	return active, warnings
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
