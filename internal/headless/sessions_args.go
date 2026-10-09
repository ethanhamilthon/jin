package headless

import (
	"errors"
	"slices"
	"strconv"
)

const sessionActionsUsage = `usage:
  jin sessions compact <id>
  jin sessions handoff <id> [--format json]
  jin sessions rewind <id> [--to <n>] [--format json]
  jin sessions undo <id>
  jin sessions context <id> [--format json]
  jin sessions reload <id>`

var sessionActions = []string{"compact", "handoff", "rewind", "undo", "context", "reload"}

// IsSessionAction tells whether the arguments after `jin sessions` name an
// action on one session, rather than list or search.
func IsSessionAction(args []string) bool {
	return len(args) > 0 && slices.Contains(sessionActions, args[0])
}

type sessionArgs struct {
	action, id, format string
	to                 int
}

func parseSessionArgs(args []string) (sessionArgs, error) {
	parsed := sessionArgs{action: args[0], format: "text"}
	for i := 1; i < len(args); i++ {
		switch arg := args[i]; {
		case arg == "--format" && i+1 < len(args):
			i++
			parsed.format = args[i]
		case arg == "--to" && i+1 < len(args):
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 {
				return parsed, errors.New("--to needs a message number from 1")
			}
			parsed.to = n
		case arg == "" || arg[0] == '-' || parsed.id != "":
			return parsed, errors.New(sessionActionsUsage)
		default:
			parsed.id = arg
		}
	}
	if parsed.id == "" || (parsed.format != "text" && parsed.format != "json") {
		return parsed, errors.New(sessionActionsUsage)
	}
	if parsed.to > 0 && parsed.action != "rewind" {
		return parsed, errors.New("--to belongs to rewind")
	}
	return parsed, nil
}
