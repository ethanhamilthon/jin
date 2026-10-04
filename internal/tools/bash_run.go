package tools

import (
	"context"
	"jin/internal/tasklog"
	"os"
	"os/exec"
	"time"
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
