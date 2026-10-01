package core

import "strings"

type RequestKind uint8

const (
	RequestPrompt RequestKind = iota
	RequestCompact
	RequestHandoff
)

// Request is one thing the user asked the agent to do. Window is the model's
// context size in tokens, 0 when unknown. NoVision is set only when the model
// is known not to accept images.
type Request struct {
	Kind                  RequestKind
	Prompt, Model, Effort string
	Window                int
	NoVision              bool
}

func (r Request) blank() bool {
	return r.Kind == RequestPrompt && strings.TrimSpace(r.Prompt) == ""
}
