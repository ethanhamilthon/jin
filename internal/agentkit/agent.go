// Package agentkit builds the agent the same way for the TUI, jin web and
// headless runs.
package agentkit

import (
	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/tasks"
	"jin/internal/tools"
)

// Mode says how long the agent lives.
type Mode int

const (
	// Interactive is a TUI or web session: background tasks outlive a run.
	Interactive Mode = iota
	// OneShot is a headless run: its tasks end with the run, and a bash
	// command past its timeout is killed.
	OneShot
)

// Spec is everything that makes up an agent before its prompts are rendered.
type Spec struct {
	Mode   Mode
	Client *provider.Client
	Names  []string
	// Dir is the working directory of the session.
	Dir string
	// Owner marks the background tasks of this agent.
	Owner string
}

// New makes the agent with its tools, working directory and background tasks.
// The system prompt comes later, from Apply.
func New(spec Spec) *core.Agent {
	agent := core.NewAgent(spec.Client, "", registry(spec))
	agent.SetWorkdir(spec.Dir)
	agent.SetBackground(tasks.Shared(), spec.Owner, spec.Mode == OneShot)
	return agent
}

func registry(spec Spec) *tools.Registry {
	if spec.Mode == OneShot {
		return tools.BuildHeadless(spec.Names)
	}
	return tools.BuildDir(spec.Names, spec.Dir)
}
