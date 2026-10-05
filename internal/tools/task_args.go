package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type taskArgs struct {
	Action, Command, Dir, ID, Text string
	Stdin                          bool
	Limit                          int
}

func parseTaskArgs(argumentsJSON string) (taskArgs, error) {
	var args taskArgs
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil {
		return args, errors.New("invalid task tool arguments")
	}
	switch args.Action {
	case "start":
		if strings.TrimSpace(args.Command) == "" {
			return args, errors.New("start needs a command")
		}
	case "check", "input", "stop":
		if args.ID == "" {
			return args, errors.New(args.Action + " needs an id")
		}
	case "list":
	default:
		return args, fmt.Errorf("unknown action %q: use start, check, input, stop or list", args.Action)
	}
	return args, nil
}
