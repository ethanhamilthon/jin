package tools

import (
	"slices"
)

// Descriptions are the one-line tool descriptions for the system prompt.
var descriptions = map[string]string{
	"read":     "read a file, optionally a line range. Read before you edit. Read a picture (png, jpeg, gif, webp, bmp) to see it.",
	"write":    "create a file or overwrite it completely.",
	"edit":     "replace an exact text match in a file. Prefer it over write for changes to existing files.",
	"bash":     "run a shell command in the working directory. A command that runs past its timeout, or while the user writes to you, keeps running as a background task (task id in the result).",
	"ask_user": "ask the user questions and wait for the answers. Use it only for real ambiguity that blocks the work.",
	"todo":     "keep a todo list for tasks with three or more steps. Send the whole list on every call; call without items to read it.",
}

// Catalog lists every tool name in the fixed display order.
func Catalog() []string {
	return []string{"read", "write", "edit", "bash", "ask_user", "todo"}
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

// Build makes a registry with the named tools. ask_user and todo need a
// store, so todos may be nil only when "todo" is not asked for.
func Build(names []string, todos TodoStore) *Registry {
	return buildDir(names, todos, NewBash(), "")
}

// BuildHeadless is Build for runs without a session that can adopt a
// background command: bash says that a command past its timeout is killed.
func BuildHeadless(names []string, todos TodoStore) *Registry {
	return buildDir(names, todos, NewBashHeadless(), "")
}

// BuildDir makes a registry whose file and bash tools use dir as their base.
func BuildDir(names []string, todos TodoStore, dir string) *Registry {
	return buildDir(names, todos, NewBash(), absoluteDir(dir))
}

func buildDir(names []string, todos TodoStore, bash Bash, dir string) *Registry {
	var list []Tool
	seen := NewSeen()
	bash.dir = dir
	for _, name := range names {
		switch name {
		case "read":
			list = append(list, Read{seen: seen, dir: dir})
		case "write":
			list = append(list, Write{seen: seen, dir: dir})
		case "edit":
			list = append(list, Edit{seen: seen, dir: dir})
		case "bash":
			list = append(list, bash)
		case "ask_user":
			list = append(list, NewAsk())
		case "todo":
			if todos != nil {
				list = append(list, NewTodo(todos))
			}
		}
	}
	return NewRegistry(list...)
}
