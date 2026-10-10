//go:build unix

package web

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
)

// serveLock serializes the Tailscale 443 check and Serve call between jin
// daemons: the check and the change must not interleave, or two daemons could
// both adopt the entry. The lock is per user.
type serveLock struct{ file *os.File }

func lockServeEntry(_ context.Context) (*serveLock, error) {
	dir := filepath.Join(os.TempDir(), "jin-remote")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(dir, "serve.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return nil, err
	}
	return &serveLock{file: file}, nil
}

func (l *serveLock) release() {
	if l == nil || l.file == nil {
		return
	}
	_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	_ = l.file.Close()
}
