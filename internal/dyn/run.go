package dyn

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// run starts one command in its own process group and waits for it, within
// the timeout. Trailing newlines are cut off its output.
func run(ctx context.Context, command string, opt Options) (string, error) {
	select {
	case slots <- struct{}{}:
		defer func() { <-slots }()
	case <-ctx.Done():
		return "", context.Canceled
	}
	runCtx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = opt.Dir
	cmd.Env = opt.Env
	setGroup(cmd)
	var out, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err := cmd.Start(); err != nil {
		return "", err
	}
	finished := make(chan error, 1)
	go func() { finished <- cmd.Wait() }()
	select {
	case err := <-finished:
		if err != nil {
			return "", failure(err, &stderr)
		}
		return strings.TrimRight(out.String(), "\r\n"), nil
	case <-runCtx.Done():
		killGroup(cmd)
		<-finished
		if ctx.Err() != nil {
			return "", context.Canceled
		}
		return "", fmt.Errorf("timed out after %s", Timeout)
	}
}

func failure(err error, stderr *bytes.Buffer) error {
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		return err
	}
	msg := strings.TrimSpace(stderr.String())
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		msg = msg[:i]
	}
	if msg == "" {
		return fmt.Errorf("exit status %d", exit.ExitCode())
	}
	return fmt.Errorf("exit status %d: %s", exit.ExitCode(), msg)
}

// sortStable keeps warnings in a steady order whatever finished first.
func sortStable(w []string) []string {
	for i := 1; i < len(w); i++ {
		for j := i; j > 0 && w[j] < w[j-1]; j-- {
			w[j], w[j-1] = w[j-1], w[j]
		}
	}
	return w
}
