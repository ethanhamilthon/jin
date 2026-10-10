//go:build !unix

package daemon

import "errors"

// Kill ends a daemon of another version. Windows needs WSL for the daemon; see
// socket_other.go.
func Kill(int) error { return errors.New("stopping a daemon of another version needs Unix") }
