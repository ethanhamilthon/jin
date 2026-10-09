package tools

import (
	"slices"
)

// Descriptions are the one-line tool descriptions for the system prompt.
var descriptions = map[string]string{
	"read":     "read a file, optionally a line range. Read before you edit. Read a picture (png, jpeg, gif, webp, bmp) to see it.",
	"grep":     "search file contents with a regular expression; matches come back as path:line:text. Prefer it to bash grep for finding where something is defined or used.",
	"write":    "create a file or overwrite it completely.",
	"edit":     "replace an exact text match in a file. Prefer it over write for changes to existing files.",
	"bash":     "run a shell command in the working directory. A command that runs past its timeout, or while the user writes to you, keeps running as a background task (task id in the result).",
	"ask_user": "ask the user questions and wait for the answers. Use it only for real ambiguity that blocks the work.",
	"task":     "run commands in the background and check, feed or stop them; results arrive as messages by themselves.",
}

// Catalog lists every tool name in the fixed display order.
func Catalog() []string {
	return []string{"read", "grep", "write", "edit", "bash", "task", "ask_user"}
}

// Describe returns the one-line description of a tool.
func Describe(name string) string { return descriptions[name] }

// Without returns the catalog minus the disabled names.
func Without(disabled []string) []string {
	var out []string
	for _, name := range Catalog() {
		if !slices.Contains(disabled, name) {
			out = append(out, name)
		}
	}
	return out
}

// Build makes a registry with the named tools.
func Build(names []string) *Registry {
	return buildDir(names, NewBash(), "")
}

// BuildHeadless is Build for runs without a session that can adopt a
// background command: bash says that a command past its timeout is killed.
func BuildHeadless(names []string) *Registry {
	return buildDir(names, NewBashHeadless(), "")
}

// BuildDir makes a registry whose file and bash tools use dir as their base.
func BuildDir(names []string, dir string) *Registry {
	return buildDir(names, NewBash(), absoluteDir(dir))
}

func buildDir(names []string, bash Bash, dir string) *Registry {
	var list []Tool
	seen := NewSeen()
	bash.dir = dir
	for _, name := range names {
		switch name {
		case "read":
			list = append(list, Read{seen: seen, dir: dir})
		case "grep":
			list = append(list, Grep{dir: dir})
		case "write":
			list = append(list, Write{seen: seen, dir: dir})
		case "edit":
			list = append(list, Edit{seen: seen, dir: dir})
		case "bash":
			list = append(list, bash)
		case "task":
			list = append(list, Task{dir: dir})
		case "ask_user":
			list = append(list, NewAsk())
		}
	}
	return NewRegistry(list...)
}
