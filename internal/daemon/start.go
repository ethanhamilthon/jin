package daemon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func Ensure(ctx context.Context, root, version string) (*Client, error) {
	client := NewClient(root)
	if _, err := client.Status(ctx); err == nil {
		return client, client.Check(ctx, version)
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	log, err := os.OpenFile(filepath.Join(root, "daemon.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(executable, "--internal-daemon")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, log, log
	isolate(cmd)
	err = cmd.Start()
	log.Close()
	if err != nil {
		return nil, err
	}
	ended := make(chan error, 1)
	go func() { ended <- cmd.Wait() }()
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		if _, err := client.Status(ctx); err == nil {
			return client, client.Check(ctx, version)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			return nil, errors.New("daemon did not start; inspect daemon.log in the jin data folder")
		case <-ended:
			// Another client may have won the startup lock.
			if _, err := client.Status(ctx); err == nil {
				return client, client.Check(ctx, version)
			}
			return nil, errors.New("daemon exited during startup; inspect daemon.log in the jin data folder")
		case <-tick.C:
		}
	}
}

func Main(ctx context.Context, args []string, version, root string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: jin daemon start|stop|status [--force]")
	}
	force := len(args) == 2 && args[0] == "stop" && args[1] == "--force"
	if len(args) != 1 && !force {
		return errors.New("usage: jin daemon start|stop|status [--force]")
	}
	client := NewClient(root)
	switch args[0] {
	case "start":
		var err error
		client, err = Ensure(ctx, root, version)
		if err != nil {
			return err
		}
	case "stop":
		if err := client.Stop(ctx, version, force); err != nil {
			return err
		}
		_, err := io.WriteString(out, "Daemon is stopping.\n")
		return err
	case "status":
	default:
		return errors.New("usage: jin daemon start|stop|status [--force]")
	}
	status, err := client.Status(ctx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Daemon %s (pid %d): working=%t, tasks=%d\n", status.Version, status.PID, status.Working, status.Tasks)
	return err
}
