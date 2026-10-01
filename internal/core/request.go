package core

import "strings"

type RequestKind uint8

const (
	RequestPrompt RequestKind = iota
	RequestCompact
	RequestHandoff
)

// Request is one thing the user asked the agent to do. Window is the model's
// context size in tokens, 0 when unknown.
type Request struct {
	Kind                  RequestKind
	Prompt, Model, Effort string
	Window                int
}

func (r Request) blank() bool {
	return r.Kind == RequestPrompt && strings.TrimSpace(r.Prompt) == ""
}
