package startup

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// commandEnv is the environment of the commands in prompts: the one of jin,
// with the folder of the running jin first in PATH so that `jin docs` and
// `jin hooks render` are this version, and JIN_DIR and JIN_SESSION_ID
// saying where the session works and which one it is.
func commandEnv(in Input) []string {
	env := slices.Clone(in.Env)
	if in.Env == nil {
		env = os.Environ()
	}
	if exe, err := os.Executable(); err == nil {
		path := filepath.Dir(exe)
		if old := lookup(env, "PATH"); old != "" {
			path += string(os.PathListSeparator) + old
		}
		env = set(env, "PATH", path)
	}
	env = set(env, "JIN_DIR", in.Dir)
	if in.SessionID != "" {
		env = set(env, "JIN_SESSION_ID", in.SessionID)
	}
	return env
}

func lookup(env []string, key string) string {
	for _, entry := range slices.Backward(env) {
		if value, ok := strings.CutPrefix(entry, key+"="); ok {
			return value
		}
	}
	return ""
}

func set(env []string, key, value string) []string {
	env = slices.DeleteFunc(env, func(entry string) bool { return strings.HasPrefix(entry, key+"=") })
	return append(env, key+"="+value)
}
