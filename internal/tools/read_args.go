package tools

import (
	"encoding/json"
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

func parseReadArgs(argumentsJSON string) (readArgs, bool) {
	var args readArgs
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil || strings.TrimSpace(args.Path) == "" {
		return readArgs{}, false
	}
	return args, true
}
