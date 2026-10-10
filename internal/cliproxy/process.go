package cliproxy

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type process struct {
	cmd      *exec.Cmd
	done     chan error
	endpoint string
	keys     keys
	version  string
}

func startProcess(ctx context.Context, root, version string, k keys) (*process, error) {
	port, err := availablePort()
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Join(root, "auth"), 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(root, "config.yaml")
	if err = writeJSON(path, configuration(root, port, k)); err != nil {
		return nil, err
	}
	cmd := exec.Command(binaryPath(root, version), "-config", path, "-local-model")
	cmd.Dir = root
	cmd.Env = processEnv(root)
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	isolate(cmd)
	if err = cmd.Start(); err != nil {
		return nil, errors.New("cannot start managed CLIProxyAPI")
	}
	p := &process{cmd: cmd, done: make(chan error, 1), endpoint: "http://127.0.0.1:" + itoa(port), keys: k, version: version}
	go func() { p.done <- cmd.Wait(); close(p.done) }()
	for {
		if _, err = p.management(ctx, "GET", "/auth-files", nil); err == nil {
			return p, nil
		}
		select {
		case <-ctx.Done():
			p.stop()
			return nil, ctx.Err()
		case <-p.done:
			return nil, errors.New("CLIProxyAPI exited before readiness")
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func (p *process) stop() {
	if p == nil {
		return
	}
	select {
	case <-p.done:
		return
	default:
	}
	terminate(p.cmd)
	select {
	case <-p.done:
	case <-time.After(5 * time.Second):
		kill(p.cmd)
		<-p.done
	}
}
