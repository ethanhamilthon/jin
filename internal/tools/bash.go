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

const bashSchema = `{"type":"function","function":{"name":"bash","description":"Run a bash command in the current working directory. When the timeout is reached, or the user writes to you while it runs, the command is not killed: it moves to the background and you get a task id (see jin async check/stop).","parameters":{"type":"object","properties":{"command":{"type":"string","description":"Shell command to execute"},"timeout":{"type":"integer","description":"Seconds to wait before the command moves to the background; must be greater than 0. Defaults to 120 if omitted."}},"required":["command"],"additionalProperties":false}}}`

type Bash struct{}

func NewBash() Bash { return Bash{} }

func (Bash) Name() string { return "bash" }

func (Bash) Schema() json.RawMessage { return json.RawMessage(bashSchema) }

func (Bash) Summary(argumentsJSON string) (string, bool) {
	command, timeout, ok := parseBashArgs(argumentsJSON)
	if !ok {
		return "", false
	}
	return command + " (" + strconv.Itoa(int(timeout.Seconds())) + "s)", true
}

func (Bash) Run(ctx context.Context, argumentsJSON string) (string, error) {
	command, timeout, ok := parseBashArgs(argumentsJSON)
	if !ok {
		return "", errors.New("invalid bash tool arguments")
	}
	return runBash(ctx, command, timeout), nil
}

// parseBashArgs validates the command and optional timeout. A timeout of
// exactly zero is rejected rather than silently treated as "no timeout" or
// "instant timeout"; omitting it entirely falls back to defaultBashTimeout.
func parseBashArgs(argumentsJSON string) (string, time.Duration, bool) {
	var args struct {
		Command string `json:"command"`
		Timeout *int   `json:"timeout"`
	}
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil {
		return "", 0, false
	}
	command := strings.TrimSpace(args.Command)
	if command == "" {
		return "", 0, false
	}
	if args.Timeout == nil {
		return command, defaultBashTimeout, true
	}
	if *args.Timeout <= 0 {
		return "", 0, false
	}
	return command, time.Duration(*args.Timeout) * time.Second, true
}
