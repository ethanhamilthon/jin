package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"jin/internal/tasks"
)

// Task starts and manages the background tasks of the session.
type Task struct{ dir string }

func (Task) Name() string { return "task" }

func (Task) Schema() json.RawMessage { return taskSchema() }

func (Task) Summary(argumentsJSON string) (string, bool) {
	args, err := parseTaskArgs(argumentsJSON)
	if err != nil {
		return "", false
	}
	switch args.Action {
	case "start":
		return "start " + args.Command, true
	case "list":
		return "list", true
	}
	return args.Action + " " + args.ID, true
}

func (t Task) Run(ctx context.Context, argumentsJSON string) (string, error) {
	args, err := parseTaskArgs(argumentsJSON)
	if err != nil {
		return "", err
	}
	bg, ok := backgroundFrom(ctx)
	if !ok {
		return "", errors.New("background tasks are not available here")
	}
	m, owner := bg.Tasks, bg.Owner
	switch args.Action {
	case "start":
		dir, err := commandDir(t.dir, args.Dir)
		if err != nil {
			return "", err
		}
		info, err := m.Start(owner, args.Command, absoluteDir(dir), args.Stdin)
		if err != nil {
			return "", err
		}
		return "started task " + info.ID + ". Its result arrives as a message when it ends; do not wait for it.", nil
	case "check":
		info, out, err := m.Check(owner, args.ID, args.Limit)
		if err != nil {
			return "", err
		}
		return describe(info) + "\n" + strings.TrimRight(out, "\n"), nil
	case "input":
		return "written to task " + args.ID, m.Input(owner, args.ID, args.Text+"\n")
	case "stop":
		return "stopping task " + args.ID, m.Stop(owner, args.ID, tasks.ByAgent)
	}
	return listTasks(m.List(owner)), nil
}
