package tools

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type readArgs struct {
	Path   string `json:"path"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

func (a readArgs) summary() string {
	if a.Offset <= 0 && a.Limit <= 0 {
		return a.Path
	}
	if a.Limit <= 0 {
		return a.Path + ":" + strconv.Itoa(a.Offset)
	}
	offset := max(a.Offset, 1)
	return a.Path + ":" + strconv.Itoa(offset) + "-" + strconv.Itoa(offset+a.Limit-1)
}

func parseReadArgs(argumentsJSON string) (readArgs, error) {
	var raw struct {
		Path   string `json:"path"`
		Offset *int   `json:"offset"`
		Limit  *int   `json:"limit"`
	}
	if json.Unmarshal([]byte(argumentsJSON), &raw) != nil || strings.TrimSpace(raw.Path) == "" {
		return readArgs{}, errors.New("invalid read tool arguments")
	}
	args := readArgs{Path: raw.Path}
	if raw.Offset != nil {
		if *raw.Offset < 1 {
			return readArgs{}, errors.New("offset must be at least 1")
		}
		args.Offset = *raw.Offset
	}
	if raw.Limit != nil {
		if *raw.Limit < 1 {
			return readArgs{}, errors.New("limit must be at least 1")
		}
		args.Limit = *raw.Limit
	}
	return args, nil
}
