package tools

import (
	"context"
	"encoding/json"
	"strconv"
	"time"
)

const maxBashOutput = 16 << 10
const defaultBashTimeout = 120 * time.Second

// Bash runs shell commands. In headless mode nothing can adopt a command that
// outlives its timeout, so it is killed and the description says so.
type Bash struct {
	headless bool
	dir      string
}

func NewBash() Bash { return Bash{} }

func NewBashHeadless() Bash { return Bash{headless: true} }

func (Bash) Name() string { return "bash" }

func (b Bash) Schema() json.RawMessage { return bashSchema(b.headless) }

func (Bash) Summary(argumentsJSON string) (string, bool) {
	args, err := parseBashArgs(argumentsJSON)
	if err != nil {
		return "", false
	}
	summary := args.command + " (" + strconv.Itoa(int(args.timeout.Seconds())) + "s)"
	if args.dir != "" {
		summary += " in " + args.dir
	}
	return summary, true
}

func (b Bash) Run(ctx context.Context, argumentsJSON string) (string, error) {
	args, err := parseBashArgs(argumentsJSON)
	if err != nil {
		return "", err
	}
	dir, err := commandDir(b.dir, args.dir)
	if err != nil {
		return "", err
	}
	return runBash(ctx, args.command, args.timeout, dir, !b.headless), nil
}
