package web

import (
	"net"
	"strconv"
)

const (
	defaultPort = 7373
	fallbacks   = 10
)

// listen opens the port the user asked for, or the default one, the next
// ten, then any free port.
func listen(port int) (net.Listener, error) {
	if port != 0 {
		return net.Listen("tcp", address(port))
	}
	for p := defaultPort; p <= defaultPort+fallbacks; p++ {
		if l, err := net.Listen("tcp", address(p)); err == nil {
			return l, nil
		}
	}
	return net.Listen("tcp", address(0))
}

func address(port int) string { return "127.0.0.1:" + strconv.Itoa(port) }
