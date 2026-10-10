package cliproxy

import (
	"context"
	"errors"
	"path/filepath"
	"time"
)

func (b *broker) update(ctx context.Context, version string) error {
	if !ValidVersion(version) {
		return errors.New("only compatible v8.0.x versions can be installed")
	}
	unlock, err := lock(b.root, "install.lock")
	if err != nil {
		return errors.New("another CLIProxyAPI installation is in progress")
	}
	defer unlock()
	current := Installed(b.root)
	if version == current.Version {
		return nil
	}
	install := Install
	if b.install != nil {
		install = b.install
	}
	if err = install(ctx, b.root, version); err != nil {
		return err
	}
	if err = checkBinary(ctx, b.root, version); err != nil {
		return err
	}
	b.mu.Lock()
	b.updating = true
	b.mu.Unlock()
	defer func() { b.mu.Lock(); b.updating = false; b.mu.Unlock(); b.notify() }()
	if err = b.waitIdle(ctx); err != nil {
		return err
	}
	b.proxy.stop()
	next, startErr := b.startVersion(version)
	if startErr == nil {
		err = writeJSON(filepath.Join(b.root, "installed.json"), Installation{Version: version, Previous: current.Version})
		if err == nil {
			b.mu.Lock()
			b.proxy = next
			b.mu.Unlock()
			return nil
		}
		next.stop()
	}
	old, recoverErr := b.startVersion(current.Version)
	if recoverErr != nil {
		return errors.New("proxy update failed and recovery failed; reinstall CLIProxyAPI in Settings")
	}
	b.mu.Lock()
	b.proxy = old
	b.mu.Unlock()
	return errors.New("proxy update failed; the previous version was restored")
}

func (b *broker) startVersion(version string) (*process, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return startProcess(ctx, b.root, version, b.keys)
}

func (c *Client) Update(ctx context.Context, version string) error {
	_, err := c.call(ctx, command{Action: "update", Version: version})
	return err
}
