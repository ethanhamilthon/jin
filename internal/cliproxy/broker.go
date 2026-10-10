package cliproxy

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type broker struct {
	mu         sync.Mutex
	operation  sync.Mutex
	root       string
	keys       keys
	proxy      *process
	clients    map[string]bool
	active     int
	operations int
	logins     map[string]string
	updating   bool
	closing    bool
	wake       chan struct{}
	install    func(context.Context, string, string) error
}

func Serve(root string) error {
	unlock, err := lock(root, "broker.lock")
	if err != nil {
		return err
	}
	defer unlock()
	installed := Installed(root)
	if installed.Version == "" {
		return errors.New("CLIProxyAPI is not installed")
	}
	k, err := loadKeys(root)
	if err != nil {
		return err
	}
	if err = socketDirectory(root); err != nil {
		return err
	}
	socket := socketPath(root)
	if info, err := os.Lstat(socket); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return errors.New("proxy control path is not a socket")
		}
		if err = os.Remove(socket); err != nil {
			return err
		}
	}
	listener, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}
	defer func() { listener.Close(); os.Remove(socket) }()
	if err = os.Chmod(socket, 0o600); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	p, err := startProcess(ctx, root, installed.Version, k)
	cancel()
	if err != nil {
		return err
	}
	b := &broker{root: root, keys: k, proxy: p, clients: map[string]bool{}, logins: map[string]string{}, wake: make(chan struct{}, 1)}
	defer func() { b.proxy.stop() }()
	server := &http.Server{Handler: b.routes(), ReadHeaderTimeout: 5 * time.Second}
	go server.Serve(listener)
	defer server.Close()
	timer := time.NewTimer(15 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-b.wake:
			b.mu.Lock()
			stop := len(b.clients) == 0 && b.active == 0 && b.operations == 0 && !b.updating
			if stop {
				b.closing = true
			}
			b.mu.Unlock()
			if stop {
				return nil
			}
		case <-timer.C:
			b.mu.Lock()
			empty := len(b.clients) == 0 && b.active == 0 && b.operations == 0 && !b.updating
			b.mu.Unlock()
			if empty {
				return nil
			}
		}
	}
}
