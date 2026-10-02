package tools

import (
	"context"
	"encoding/json"
	"slices"
)

type Tool interface {
	Name() string
	Schema() json.RawMessage
	Summary(argumentsJSON string) (string, bool)
	Run(ctx context.Context, argumentsJSON string) (string, error)
}

// ImageTool is a Tool that can also return pictures for the model to see.
type ImageTool interface {
	Tool
	RunImages(ctx context.Context, argumentsJSON string) (string, []Image, error)
}

type Registry struct {
	tools map[string]Tool
	order []string
}

func NewRegistry(list ...Tool) *Registry {
	r := &Registry{tools: make(map[string]Tool, len(list))}
	for _, t := range list {
		r.tools[t.Name()] = t
		r.order = append(r.order, t.Name())
	}
	return r
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) Names() []string {
	return append([]string(nil), r.order...)
}

func (r *Registry) SchemaJSON() json.RawMessage {
	if len(r.order) == 0 {
		return nil
	}
	parts := make([]json.RawMessage, 0, len(r.order))
	for _, name := range r.order {
		parts = append(parts, r.tools[name].Schema())
	}
	encoded, _ := json.Marshal(parts)
	return encoded
}

// Descriptions are the one-line tool descriptions for the system prompt.
var descriptions = map[string]string{
	"read":     "read a file, optionally a line range. Read before you edit.",
	"write":    "create a file or overwrite it completely.",
	"edit":     "replace an exact text match in a file. Prefer it over write for changes to existing files.",
	"bash":     "run a shell command in the working directory.",
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
	var list []Tool
	for _, name := range names {
		switch name {
		case "read":
			list = append(list, NewRead())
		case "write":
			list = append(list, NewWrite())
		case "edit":
			list = append(list, NewEdit())
		case "bash":
			list = append(list, NewBash())
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
