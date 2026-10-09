package web

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Options are the flags of `jin web`.
type Options struct {
	Port   int
	NoOpen bool
	Cwd    string
	Help   bool
	Hosts  []string
}

const Usage = `usage: jin web [--port N] [--no-open] [--cwd DIR] [--allow-host NAME]

  --port N     listen on this port; a busy port is an error
               (default 7373, then 7374-7383, then any free port)
  --no-open    print the URL instead of opening the browser
  --cwd DIR    the project the page opens first (default: this directory)
  --allow-host NAME
               also accept this host name over https, for a reverse proxy such
               as "tailscale serve" (repeatable); jin still listens on 127.0.0.1
`

// ParseArgs reads the flags after `jin web`.
func ParseArgs(args []string) (Options, error) {
	var opt Options
	for i := 0; i < len(args); i++ {
		value := func() (string, error) {
			if i+1 >= len(args) {
				return "", fmt.Errorf("%s needs a value", args[i])
			}
			i++
			return args[i], nil
		}
		switch args[i] {
		case "--port":
			v, err := value()
			if err != nil {
				return opt, err
			}
			port, err := strconv.Atoi(v)
			if err != nil || port < 1 || port > 65535 {
				return opt, errors.New("--port must be a number from 1 to 65535")
			}
			opt.Port = port
		case "--cwd":
			v, err := value()
			if err != nil {
				return opt, err
			}
			opt.Cwd = v
		case "--allow-host":
			v, err := value()
			if err != nil {
				return opt, err
			}
			if v == "" || strings.ContainsAny(v, "/ ") {
				return opt, errors.New("--allow-host needs a bare host name, for example my-mac.tailnet.ts.net")
			}
			opt.Hosts = append(opt.Hosts, v)
		case "--no-open":
			opt.NoOpen = true
		case "--help", "-h":
			opt.Help = true
		default:
			return opt, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return opt, nil
}
