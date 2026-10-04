package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"jin/internal/provider"
)

const nestedAgentsLimit = 8 << 10

// nestedSent remembers, per agent, which subdirectory AGENTS.md files the
// model already got. The agent outlives compaction, so the set does too.
var nestedSent sync.Map // *Agent -> *sentFiles

type sentFiles struct {
	mu    sync.Mutex
	files map[string]bool
}

func pathOf(call provider.ToolCall) (string, bool) {
	switch call.Function.Name {
	case "read", "edit", "write":
		var args struct {
			Path string `json:"path"`
		}
		return args.Path, json.Unmarshal([]byte(call.Function.Arguments), &args) == nil && args.Path != ""
	}
	return "", false
}

// nestedAgents returns the text to append to the result of a read, edit or
// write call: the AGENTS.md files of the folders it reached for the first time.
func (a *Agent) nestedAgents(call provider.ToolCall, result string) string {
	path, ok := pathOf(call)
	if !ok || strings.HasPrefix(result, "Error:") {
		return ""
	}
	workdir := a.workdir
	if workdir == "" {
		var err error
		workdir, err = os.Getwd()
		if err != nil {
			return ""
		}
	}
	entry, _ := nestedSent.LoadOrStore(a, &sentFiles{files: map[string]bool{}})
	sent := entry.(*sentFiles)
	sent.mu.Lock()
	files := NearestUnseen(workdir, path, sent.files)
	sent.mu.Unlock()
	for i := range files {
		if len(files[i].Content) > nestedAgentsLimit {
			total := len(files[i].Content)
			cut := strings.ToValidUTF8(files[i].Content[:nestedAgentsLimit], "")
			files[i].Content = cut + fmt.Sprintf("\n[truncated: %d of %d bytes shown]", len(cut), total)
		}
	}
	if len(files) == 0 {
		return ""
	}
	return "\n\n" + renderContext(files)
}
