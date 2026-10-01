package tools

import (
	"context"
	"encoding/json"
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
	parts := make([]json.RawMessage, 0, len(r.order))
	for _, name := range r.order {
		parts = append(parts, r.tools[name].Schema())
	}
	encoded, _ := json.Marshal(parts)
	return encoded
}
