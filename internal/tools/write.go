package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const writeSchema = `{"type":"function","function":{"name":"write","description":"Create or overwrite a file with the given content","parameters":{"type":"object","properties":{"path":{"type":"string","description":"File path to write"},"content":{"type":"string","description":"Full file content"}},"required":["path","content"],"additionalProperties":false}}}`

type Write struct{ seen *Seen }

func NewWrite() Write { return Write{} }

// NewWriteSeen shares seen with the other file tools of a session.
func NewWriteSeen(seen *Seen) Write { return Write{seen: seen} }

func (Write) Name() string { return "write" }

func (Write) Schema() json.RawMessage { return json.RawMessage(writeSchema) }

func (Write) Summary(argumentsJSON string) (string, bool) {
	args, ok := parseWriteArgs(argumentsJSON)
	if !ok {
		return "", false
	}
	return args.Path, true
}

func (w Write) Run(ctx context.Context, argumentsJSON string) (string, error) {
	args, ok := parseWriteArgs(argumentsJSON)
	if !ok {
		return "", errors.New("invalid write tool arguments")
	}
	if err := w.seen.Check(args.Path); err != nil {
		return "", err
	}
	before, existed, err := currentContent(args.Path)
	if err != nil {
		return "", err
	}
	if dir := filepath.Dir(args.Path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
	}
	if err := os.WriteFile(args.Path, []byte(args.Content), 0o644); err != nil {
		return "", err
	}
	w.seen.Remember(args.Path)
	reportChange(ctx, Change{Path: args.Path, Existed: existed, Before: before, After: args.Content})
	return "Wrote " + strconv.Itoa(len(args.Content)) + " bytes to " + args.Path, nil
}

type writeArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func parseWriteArgs(argumentsJSON string) (writeArgs, bool) {
	var args writeArgs
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil || strings.TrimSpace(args.Path) == "" {
		return writeArgs{}, false
	}
	return args, true
}
