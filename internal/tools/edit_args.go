package tools

import (
	"encoding/json"
	"errors"
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
	Path       string
	OldString  string
	NewString  string
	ReplaceAll bool
}

func parseEditArgs(argumentsJSON string) (editArgs, error) {
	var raw struct {
		Path       *string `json:"path"`
		OldString  *string `json:"old_string"`
		NewString  *string `json:"new_string"`
		ReplaceAll bool    `json:"replace_all"`
	}
	if json.Unmarshal([]byte(argumentsJSON), &raw) != nil {
		return editArgs{}, errors.New("invalid edit tool arguments")
	}
	switch {
	case raw.Path == nil || strings.TrimSpace(*raw.Path) == "":
		return editArgs{}, errMissingField("path")
	case raw.OldString == nil:
		return editArgs{}, errMissingField("old_string")
	case *raw.OldString == "":
		return editArgs{}, errors.New("old_string must not be empty")
	case raw.NewString == nil:
		return editArgs{}, errMissingField("new_string")
	}
	return editArgs{*raw.Path, *raw.OldString, *raw.NewString, raw.ReplaceAll}, nil
}

func errMissingField(name string) error {
	return errors.New("missing required field: " + name)
}
