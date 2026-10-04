//go:build unix

package tools

import (
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestReadRefusesFIFOWithoutHanging(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pipe")
	if err := syscall.Mkfifo(path, 0o644); err != nil {
		t.Skip("mkfifo unavailable")
	}
	done := make(chan error, 1)
	go func() { _, err := runRead(t, map[string]any{"path": path}); done <- err }()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "not a regular file") {
			t.Fatalf("err=%v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("read of a FIFO hung")
	}
}
