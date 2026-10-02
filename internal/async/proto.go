// Package async runs background commands for agents. One daemon per user
// supervises the processes, so a task outlives the TUI and the agent that
// started it. When a task ends, its result is queued as an event for the
// session, and the TUI of that directory hands it to the agent.
package async

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"jin/internal/paths"
	"jin/internal/tasklog"
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
	Op        string   `json:"op"`
	Session   string   `json:"session,omitempty"`
	Cwd       string   `json:"cwd,omitempty"`
	Command   string   `json:"command,omitempty"`
	Env       []string `json:"env,omitempty"`
	ID        string   `json:"id,omitempty"`
	Text      string   `json:"text,omitempty"`
	NoNewline bool     `json:"no_newline,omitempty"`
	By        string   `json:"by,omitempty"`

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

// ResultText wraps what the agent should see when a task ends. A negative
// exit leaves the exit attribute out. The closing tag inside the output is
// broken up, so output can never close the wrapper and pose as the user.
func ResultText(id, status string, exit int, body string) string {
	body = strings.TrimRight(body, " \t\r\n")
	if body == "" {
		body = "(no output)"
	}
	body = strings.ReplaceAll(body, closeTag, `<\/async-task-result`)
	attrs := fmt.Sprintf(`id=%q status=%q`, id, status)
	if exit >= 0 {
		attrs += fmt.Sprintf(` exit="%d"`, exit)
	}
	return "<async-task-result " + attrs + ">\n" + body + "\n</async-task-result>"
}

var attrPattern = regexp.MustCompile(`(\w+)="([^"]*)"`)

// Summary turns a result message into the lines the chat shows: a headline
// and the first lines of the output. ok is false for any other text.
func Summary(text string) (summary string, ok bool) {
	rest, found := strings.CutPrefix(text, "<async-task-result ")
	if !found {
		return "", false
	}
	header, body, _ := strings.Cut(rest, ">\n")
	attrs := map[string]string{}
	for _, m := range attrPattern.FindAllStringSubmatch(header, -1) {
		attrs[m[1]] = m[2]
	}
	line := "async task " + attrs["id"] + " " + attrs["status"]
	if exit, has := attrs["exit"]; has {
		line += " (exit " + exit + ")"
	}
	body = strings.TrimSuffix(strings.TrimRight(body, "\n"), "</async-task-result>")
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	if len(lines) > summaryLines {
		lines = append(lines[:summaryLines], "...")
	}
	return line + "\n" + strings.Join(lines, "\n"), true
}

// LogPath is the file that holds the output of a task.
func LogPath(id string) (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return dir + string(os.PathSeparator) + id + ".log", nil
}

// Tail reads the last limit characters of a file; limit <= 0 reads all of
// it. The second result tells that the start was cut off.
func Tail(path string, limit int) (string, bool, error) { return tasklog.Tail(path, limit) }
