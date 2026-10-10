//go:build unix

package daemon

import (
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func socketPath(root string) string {
	if canonical, err := filepath.EvalSymlinks(root); err == nil {
		root = canonical
	}
	hash := sha256.Sum256([]byte(filepath.Clean(root)))
	return filepath.Join(fmt.Sprintf("/tmp/jin-daemon-%d", os.Getuid()), fmt.Sprintf("%x.sock", hash[:16]))
}

func listen(root string) (net.Listener, func(), error) {
	dir := filepath.Dir(socketPath(root))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, nil, err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return nil, nil, err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !info.IsDir() || info.Mode().Perm() != 0o700 || !ok || stat.Uid != uint32(os.Getuid()) {
		return nil, nil, fmt.Errorf("unsafe daemon control directory")
	}
	file, err := os.OpenFile(filepath.Join(root, "daemon.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, nil, err
	}
	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("daemon is already running: %w", err)
	}
	unlock := func() { file.Close() }
	socket := socketPath(root)
	if info, err := os.Lstat(socket); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			unlock()
			return nil, nil, fmt.Errorf("daemon control path is not a socket")
		}
		if err = os.Remove(socket); err != nil {
			unlock()
			return nil, nil, err
		}
	} else if !os.IsNotExist(err) {
		unlock()
		return nil, nil, err
	}
	listener, err := net.Listen("unix", socket)
	if err != nil {
		unlock()
		return nil, nil, err
	}
	cleanup := func() { listener.Close(); os.Remove(socket); unlock() }
	if err = os.Chmod(socket, 0o600); err != nil {
		cleanup()
		return nil, nil, err
	}
	return listener, cleanup, nil
}

func isolate(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }
