package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"jin/internal/tasklog"
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

// maxLog bounds the log of a command that has moved to the background; the
// last keepLog bytes stay when it is exceeded.
const (
	maxLog  = 64 << 20
	keepLog = 16 << 20
)

// runBash runs a command and waits for it, until it ends, the timeout is
// reached, or the user asks for the background. A command that is still
// running at that point becomes an async task instead of being killed, when
// the session can adopt tasks; otherwise it is killed as before.
func runBash(ctx context.Context, command string, timeout time.Duration) string {
	background, canAdopt := backgroundFrom(ctx)
	files, err := tasklog.New()
	if err != nil {
		return "[command failed: " + err.Error() + "]"
	}
	logFile, err := os.OpenFile(files.Log, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		tasklog.Remove(files.Log)
		return "[command failed: " + err.Error() + "]"
	}
	defer logFile.Close()

	cmd := exec.Command("bash", "-c", command)
	setGroup(cmd)
	cmd.Stdout, cmd.Stderr = logFile, logFile
	cmd.Stdin = nil
	if err := cmd.Start(); err != nil {
		tasklog.Remove(files.Log)
		return "[command failed: " + err.Error() + "]"
	}
	finished := make(chan error, 1)
	go func() { finished <- cmd.Wait() }()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	var detach <-chan struct{}
	if canAdopt {
		detach = background.Detach
	}
	select {
	case err := <-finished:
		return finishBash(cmd, files, err, false)
	case <-timer.C:
		return moveOn(cmd, files, command, background, canAdopt, finished, "timed out after "+timeout.String())
	case <-detach:
		return moveOn(cmd, files, command, background, canAdopt, finished, "moved to the background at your user's request")
	case <-ctx.Done():
		killProcessGroup(cmd)
		err := <-finished
		return finishBash(cmd, files, err, false)
	}
}

// moveOn hands a running command to the daemon, or kills it when that is not
// possible. why is the sentence that tells the agent what happened.
func moveOn(cmd *exec.Cmd, files tasklog.Files, command string, background Background, canAdopt bool, finished chan error, why string) string {
	if !canAdopt {
		killProcessGroup(cmd)
		err := <-finished
		return finishBash(cmd, files, err, true)
	}
	// The exit code is written by jin when the process ends, because the
	// daemon is not its parent and cannot wait for it.
	go func() {
		err := <-finished
		writeExit(files.Exit, cmd, err)
	}()
	id, err := background.Adopt(Adoption{PID: cmd.Process.Pid, PGID: cmd.Process.Pid, Command: command, Log: files.Log, Exit: files.Exit})
	if err != nil {
		killProcessGroup(cmd)
		tasklog.Remove(files.Log, files.Exit)
		return readLog(files.Log) + "\n[command " + why + "; it could not move to the background (" + err.Error() + ") and was stopped]"
	}
	return readLog(files.Log) + "\n[command " + why + ". It keeps running as background task " + id +
		". Its result will arrive as a message when it ends. Use `jin async check --id " + id +
		" --limit 2000` to look at it and `jin async stop --id " + id + "` to stop it.]"
}

// finishBash builds the result of a command that ended while we waited.
// killed is set when jin ended it because its time was up and nothing could
// adopt it.
func finishBash(cmd *exec.Cmd, files tasklog.Files, err error, killed bool) string {
	result, truncated, _ := tasklog.Head(files.Log, maxBashOutput)
	tasklog.Remove(files.Log, files.Exit)
	if truncated {
		result += "\n[output truncated]"
	}
	switch {
	case killed:
		result += "\n[command timed out]"
	case err != nil:
		result += "\n[command failed: " + err.Error() + "]"
	case result == "":
		result = "[command completed with no output]"
	}
	rememberGroup(cmd)
	return result
}

// readLog is the start of the output of a command that moved on.
func readLog(path string) string {
	text, truncated, _ := tasklog.Head(path, maxBashOutput)
	if truncated {
		text += "\n[output truncated; read the rest with jin async check]"
	}
	return text
}

// writeExit records how a command ended, for the daemon. It writes to a
// temporary name first, so the daemon never reads a half-written file.
func writeExit(path string, cmd *exec.Cmd, err error) {
	code := 0
	if cmd.ProcessState != nil {
		code = cmd.ProcessState.ExitCode()
		if code < 0 {
			code = signalCode(cmd)
		}
	} else if err != nil {
		code = 1
	}
	tmp := path + ".tmp"
	if os.WriteFile(tmp, []byte(strconv.Itoa(code)+"\n"), 0o600) == nil {
		_ = os.Rename(tmp, path)
	}
}
