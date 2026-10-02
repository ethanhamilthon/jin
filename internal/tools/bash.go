package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const maxBashOutput = 16 << 10
const defaultBashTimeout = 120 * time.Second

const bashSchema = `{"type":"function","function":{"name":"bash","description":"Run a bash command in the current working directory","parameters":{"type":"object","properties":{"command":{"type":"string","description":"Shell command to execute"},"timeout":{"type":"integer","description":"Timeout in seconds; must be greater than 0. Defaults to 120 if omitted."}},"required":["command"],"additionalProperties":false}}}`

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

func runBash(ctx context.Context, command string, timeout time.Duration) string {
	commandCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(commandCtx, "bash", "-c", command)
	cmd.WaitDelay = time.Second
	killGroup(cmd)
	var output boundedOutput
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	rememberGroup(cmd)
	result := output.buffer.String()
	if output.truncated {
		result += "\n[output truncated]"
	}
	if commandCtx.Err() == context.DeadlineExceeded {
		result += "\n[command timed out]"
	} else if err != nil {
		result += "\n[command failed: " + err.Error() + "]"
	} else if result == "" {
		result = "[command completed with no output]"
	}
	return result
}
