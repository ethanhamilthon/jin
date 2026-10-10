package cliproxy

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

type keys struct {
	Client     string `json:"client"`
	Management string `json:"management"`
	Control    string `json:"control"`
}

func loadKeys(root string) (keys, error) {
	var k keys
	path := filepath.Join(root, "keys.json")
	if err := readJSON(path, &k); err == nil {
		if len(k.Client) != 64 || len(k.Management) != 64 || len(k.Control) != 64 {
			return keys{}, fmt.Errorf("invalid managed proxy keys")
		}
		return k, nil
	} else if !os.IsNotExist(err) {
		return k, err
	}
	for _, value := range []*string{&k.Client, &k.Management, &k.Control} {
		var data [32]byte
		if _, err := rand.Read(data[:]); err != nil {
			return k, err
		}
		*value = hex.EncodeToString(data[:])
	}
	return k, writeJSON(path, k)
}

func configuration(root string, port int, k keys) map[string]any {
	return map[string]any{
		"host": "127.0.0.1", "port": port, "auth-dir": filepath.Join(root, "auth"), "api-keys": []string{k.Client},
		"remote-management":  map[string]any{"allow-remote": false, "secret-key": k.Management, "disable-control-panel": true, "disable-auto-update-panel": true},
		"force-model-prefix": true, "request-retry": 0, "max-retry-credentials": 1,
		"quota-exceeded": map[string]any{"switch-project": false, "switch-preview-model": false, "antigravity-credits": false},
		"plugins":        map[string]any{"enabled": false}, "logging-to-file": false, "request-log": false,
		"upstream": map[string]any{"claude": map[string]any{"disable-claude-cloak-mode": true, "disable-cloaking-model-list": true}, "codex": map[string]any{"disable-codex-cloaking": true}},
	}
}

func availablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}
