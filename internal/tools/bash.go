package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

const maxBashOutput = 16 << 10
const defaultBashTimeout = 120 * time.Second

// Bash runs shell commands. In headless mode nothing can adopt a command that
// outlives its timeout, so it is killed and the description says so.
type Bash struct {
	headless bool
	dir      string
}

func NewBash() Bash { return Bash{} }

func NewBashHeadless() Bash { return Bash{headless: true} }

func (Bash) Name() string { return "bash" }

func (b Bash) Schema() json.RawMessage { return bashSchema(b.headless) }

func (Bash) Summary(argumentsJSON string) (string, bool) {
	command, timeout, err := parseBashArgs(argumentsJSON)
	if err != nil {
		return "", false
	}
	return command + " (" + strconv.Itoa(int(timeout.Seconds())) + "s)", true
}

func (b Bash) Run(ctx context.Context, argumentsJSON string) (string, error) {
	command, timeout, err := parseBashArgs(argumentsJSON)
	if err != nil {
		return "", err
	}
	return runBashInDir(ctx, command, timeout, b.dir), nil
}

// parseBashArgs validates the command and optional timeout. A timeout below
// one second is an error; omitting it falls back to defaultBashTimeout.
func parseBashArgs(argumentsJSON string) (string, time.Duration, error) {
	var args struct {
		Command string `json:"command"`
		Timeout *int   `json:"timeout"`
	}
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil {
		return "", 0, errors.New("invalid bash tool arguments")
	}
	command := strings.TrimSpace(args.Command)
	if command == "" {
		return "", 0, errors.New("invalid bash tool arguments: command must not be empty")
	}
	if args.Timeout == nil {
		return command, defaultBashTimeout, nil
	}
	if *args.Timeout < 1 {
		return "", 0, errors.New("timeout must be at least 1 second")
	}
	return command, time.Duration(*args.Timeout) * time.Second, nil
}
