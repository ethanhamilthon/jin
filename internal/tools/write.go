package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
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
	args, err := parseWriteArgs(argumentsJSON)
	if err != nil {
		return "", false
	}
	return args.Path, true
}

func (w Write) Run(ctx context.Context, argumentsJSON string) (string, error) {
	args, err := parseWriteArgs(argumentsJSON)
	if err != nil {
		return "", err
	}
	if err := w.seen.Check(args.Path); err != nil {
		return "", err
	}
	target, err := writeTarget(args.Path)
	if err != nil {
		return "", err
	}
	before, existed, err := currentContent(target)
	if err != nil {
		return "", err
	}
	if dir := parentOf(target); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
	}
	if err := writeFileAtomic(target, []byte(args.Content), 0o644); err != nil {
		return "", err
	}
	w.seen.Remember(args.Path)
	reportChange(ctx, Change{Path: target, Existed: existed, Before: before, After: args.Content})
	return "Wrote " + strconv.Itoa(len(args.Content)) + " bytes to " + args.Path, nil
}

type writeArgs struct {
	Path    string
	Content string
}

func parseWriteArgs(argumentsJSON string) (writeArgs, error) {
	var raw struct {
		Path    *string `json:"path"`
		Content *string `json:"content"`
	}
	if json.Unmarshal([]byte(argumentsJSON), &raw) != nil {
		return writeArgs{}, errors.New("invalid write tool arguments")
	}
	if raw.Path == nil || strings.TrimSpace(*raw.Path) == "" {
		return writeArgs{}, errMissingField("path")
	}
	if raw.Content == nil {
		return writeArgs{}, errMissingField("content")
	}
	return writeArgs{Path: *raw.Path, Content: *raw.Content}, nil
}
