package tools

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type bashArgs struct {
	command string
	timeout time.Duration
	dir     string
}

// parseBashArgs validates the command, the optional timeout and the optional
// dir. A timeout below one second is an error; omitting it falls back to
// defaultBashTimeout.
func parseBashArgs(argumentsJSON string) (bashArgs, error) {
	var raw struct {
		Command string `json:"command"`
		Timeout *int   `json:"timeout"`
		Dir     string `json:"dir"`
	}
	if json.Unmarshal([]byte(argumentsJSON), &raw) != nil {
		return bashArgs{}, errors.New("invalid bash tool arguments")
	}
	args := bashArgs{command: strings.TrimSpace(raw.Command), timeout: defaultBashTimeout, dir: strings.TrimSpace(raw.Dir)}
	if args.command == "" {
		return bashArgs{}, errors.New("invalid bash tool arguments: command must not be empty")
	}
	if raw.Timeout != nil {
		if *raw.Timeout < 1 {
			return bashArgs{}, errors.New("timeout must be at least 1 second")
		}
		args.timeout = time.Duration(*raw.Timeout) * time.Second
	}
	return args, nil
}

// commandDir is where a command starts: dir resolved against the session's
// base directory, or the base itself when dir is empty.
func commandDir(base, dir string) (string, error) {
	if dir == "" {
		return base, nil
	}
	path := dir
	if !filepath.IsAbs(path) {
		path = filepath.Join(absoluteDir(base), dir)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", errors.New("dir " + dir + " does not exist")
	}
	if !info.IsDir() {
		return "", errors.New("dir " + dir + " is not a directory")
	}
	return path, nil
}
