//go:build !unix

package daemon

import (
	"errors"
	"net"
	"os/exec"
	"path/filepath"
)

func socketPath(root string) string { return filepath.Join(root, "daemon.sock") }
func listen(string) (net.Listener, func(), error) {
	return nil, nil, errors.New("the daemon requires Unix sockets; use WSL on Windows")
}
func isolate(cmd *exec.Cmd) {}
