//go:build unix

package async

import (
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"jin/internal/store"
)

const (
	idleExit   = 10 * time.Minute
	killGrace  = 3 * time.Second
	connLimit  = 30 * time.Second
	requestMax = 8 << 20
)

type daemon struct {
	db      *store.DB
	version string
	mu      sync.Mutex
	tasks   map[string]*task
	active  time.Time
	done    chan struct{}
}

// Serve runs the daemon until it has been idle for ten minutes. A second
// daemon finds the lock taken and returns at once.
func Serve(ctx context.Context, db *store.DB, version string) error {
	dir, err := dataDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	lock, err := lockPath()
	if err != nil {
		return err
	}
	lockFile, err := os.OpenFile(lock, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer lockFile.Close()
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil
		}
		return err
	}
	sock, err := socketPath()
	if err != nil {
		return err
	}
	_ = os.Remove(sock)
	listener, err := net.Listen("unix", sock)
	if err != nil {
		return err
	}
	defer os.Remove(sock)
	if err := os.Chmod(sock, 0o600); err != nil {
		listener.Close()
		return err
	}
	d := &daemon{db: db, version: version, tasks: map[string]*task{}, active: time.Now(), done: make(chan struct{})}
	d.recover()
	cleanOldFiles(dir)
	go d.accept(listener)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			listener.Close()
			return nil
		case <-d.done:
			listener.Close()
			return nil
		case <-ticker.C:
			if d.idleFor() > idleExit {
				listener.Close()
				return nil
			}
		}
	}
}
