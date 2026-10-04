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
	args, err := parseEditArgs(argumentsJSON)
	if err != nil {
		return "", false
	}
	if lines, ok := matchLineRange(args.Path, args.OldString); ok {
		return args.Path + ":" + lines, true
	}
	return args.Path, true
}

func (e Edit) Run(ctx context.Context, argumentsJSON string) (string, error) {
	args, err := parseEditArgs(argumentsJSON)
	if err != nil {
		return "", err
	}
	if err := e.seen.Check(args.Path); err != nil {
		return "", err
	}
	target, err := writeTarget(args.Path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(target)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(target)
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
	if err := writeFileAtomic(target, []byte(updated), info.Mode().Perm()); err != nil {
		return "", err
	}
	e.seen.Remember(args.Path)
	reportChange(ctx, Change{Path: target, Existed: true, Before: content, After: updated})
	replaced := 1
	if args.ReplaceAll {
		replaced = count
	}
	return "Replaced " + strconv.Itoa(replaced) + " occurrence(s) in " + args.Path, nil
}
