package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"jin/internal/paths"
)

type debugState struct {
	mu                          sync.Mutex
	path                        string
	previous                    map[string][]string
	input, cached, hits, misses int
}

func newDebugState() *debugState {
	value := strings.TrimSpace(os.Getenv("JIN_DEBUG"))
	if value == "" || value == "0" {
		return nil
	}
	path := value
	if value == "1" {
		path, _ = paths.Global("debug", "provider-"+processID()+".jsonl")
	}
	return &debugState{path: path, previous: make(map[string][]string)}
}

// Debug writes metadata only. It never writes prompts, tool arguments or headers.
func (c *Client) Debug(event string, fields map[string]any) {
	if c.debug == nil {
		return
	}
	c.debug.mu.Lock()
	defer c.debug.mu.Unlock()
	cfg := c.Config()
	fields["event"], fields["time"] = event, time.Now().UTC().Format(time.RFC3339Nano)
	fields["pid"], fields["kind"] = os.Getpid(), cfg.Kind
	data, err := json.Marshal(fields)
	if err != nil || os.MkdirAll(filepath.Dir(c.debug.path), 0o700) != nil {
		return
	}
	if info, err := os.Lstat(c.debug.path); err == nil && !info.Mode().IsRegular() {
		return
	}
	file, err := os.OpenFile(c.debug.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer file.Close()
	if file.Chmod(0o600) != nil {
		return
	}
	_, _ = file.WriteString(redact(string(data), cfg.APIKey) + "\n")
}
