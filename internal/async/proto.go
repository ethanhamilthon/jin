// Package async runs background commands for agents. One daemon per user
// supervises the processes, so a task outlives the TUI and the agent that
// started it. When a task ends, its result is queued as an event for the
// session, and the TUI of that directory hands it to the agent.
package async

import (
	"errors"

	"jin/internal/paths"
)

// Operations of the socket protocol: one JSON request, one JSON response.
const (
	opHello    = "hello"
	opRun      = "run"
	opInput    = "input"
	opStop     = "stop"
	opAdopt    = "adopt"
	opShutdown = "shutdown"
)

// StoppedByUser and StoppedByAgent say who stopped a task.
const (
	StoppedByUser  = "user"
	StoppedByAgent = "agent"
)

type request struct {
	Op          string   `json:"op"`
	Session     string   `json:"session,omitempty"`
	Cwd         string   `json:"cwd,omitempty"`
	CwdExplicit bool     `json:"cwd_explicit,omitempty"`
	Command     string   `json:"command,omitempty"`
	Env         []string `json:"env,omitempty"`
	ID          string   `json:"id,omitempty"`
	Text        string   `json:"text,omitempty"`
	NoNewline   bool     `json:"no_newline,omitempty"`
	By          string   `json:"by,omitempty"`
	Stdin       bool     `json:"stdin,omitempty"`

	// Adopt: a process started by the bash tool that moves to the background.
	PID  int    `json:"pid,omitempty"`
	PGID int    `json:"pgid,omitempty"`
	Log  string `json:"log,omitempty"`
	Exit string `json:"exit,omitempty"`
}

type response struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	Version string `json:"version,omitempty"`
	// Idle is set on hello: the daemon runs no task and may be replaced.
	Idle bool   `json:"idle,omitempty"`
	ID   string `json:"id,omitempty"`
}

var errUnsupported = errors.New("async is not supported on this OS")

const (
	// eventTail is how much of a task's output goes into its result event.
	eventTail = 8000

	closeTag = "</async-task-result"

	// summaryLines is how many lines of a result the chat shows.
	summaryLines = 6
)

func dataDir() (string, error) { return paths.Global("async") }

func socketPath() (string, error) { return paths.Global("async.sock") }

func lockPath() (string, error) { return paths.Global("async.lock") }
