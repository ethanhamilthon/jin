package core

type ContextKind string

const ContextProject ContextKind = "project"

const agentsFile = "AGENTS.md"

// ContextFile is the AGENTS.md of the working directory. Jin only reports it
// (the intro lists it); the system prompt file reads it with a command.
type ContextFile struct {
	Path    string
	Kind    ContextKind
	Content string
}
