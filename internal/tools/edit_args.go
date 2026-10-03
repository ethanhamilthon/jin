package tools

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

// matchLineRange locates old_string's 1-based line range in path's current
// content, for display only. A miss (file unreadable, text already
// changed) just means no line range is shown.
func matchLineRange(path, oldString string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	content := string(data)
	idx := strings.Index(content, oldString)
	if idx < 0 {
		return "", false
	}
	start := 1 + strings.Count(content[:idx], "\n")
	end := start + strings.Count(oldString, "\n")
	if end == start {
		return strconv.Itoa(start), true
	}
	return strconv.Itoa(start) + "-" + strconv.Itoa(end), true
}

type editArgs struct {
	Path       string `json:"path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all"`
}

func parseEditArgs(argumentsJSON string) (editArgs, bool) {
	var args editArgs
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil || strings.TrimSpace(args.Path) == "" || args.OldString == "" {
		return editArgs{}, false
	}
	return args, true
}
