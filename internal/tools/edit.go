package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
)

const editSchema = `{"type":"function","function":{"name":"edit","description":"Replace an exact text match in a file with new text","parameters":{"type":"object","properties":{"path":{"type":"string","description":"File path to edit"},"old_string":{"type":"string","description":"Exact text to find"},"new_string":{"type":"string","description":"Text to replace it with"},"replace_all":{"type":"boolean","description":"Replace every occurrence instead of requiring exactly one"}},"required":["path","old_string","new_string"],"additionalProperties":false}}}`

type Edit struct{ seen *Seen }

func NewEdit() Edit { return Edit{} }

// NewEditSeen shares seen with the other file tools of a session.
func NewEditSeen(seen *Seen) Edit { return Edit{seen: seen} }

func (Edit) Name() string { return "edit" }

func (Edit) Schema() json.RawMessage { return json.RawMessage(editSchema) }

func (Edit) Summary(argumentsJSON string) (string, bool) {
	args, ok := parseEditArgs(argumentsJSON)
	if !ok {
		return "", false
	}
	if lines, ok := matchLineRange(args.Path, args.OldString); ok {
		return args.Path + ":" + lines, true
	}
	return args.Path, true
}

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

func (e Edit) Run(_ context.Context, argumentsJSON string) (string, error) {
	args, ok := parseEditArgs(argumentsJSON)
	if !ok {
		return "", errors.New("invalid edit tool arguments")
	}
	if err := e.seen.Check(args.Path); err != nil {
		return "", err
	}
	info, err := os.Stat(args.Path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(args.Path)
	if err != nil {
		return "", err
	}
	content := string(data)
	count := strings.Count(content, args.OldString)
	switch {
	case count == 0:
		return "", errors.New("old_string not found in " + args.Path)
	case count > 1 && !args.ReplaceAll:
		return "", errors.New("old_string matches " + strconv.Itoa(count) + " places in " + args.Path + "; add more context or set replace_all")
	}
	limit := 1
	if args.ReplaceAll {
		limit = -1
	}
	updated := strings.Replace(content, args.OldString, args.NewString, limit)
	if err := os.WriteFile(args.Path, []byte(updated), info.Mode()); err != nil {
		return "", err
	}
	e.seen.Remember(args.Path)
	replaced := 1
	if args.ReplaceAll {
		replaced = count
	}
	return "Replaced " + strconv.Itoa(replaced) + " occurrence(s) in " + args.Path, nil
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
