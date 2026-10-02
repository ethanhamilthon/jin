package ui

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"jin/internal/async"
	"jin/internal/core"
)

const asyncTailChars = 4000

// openAsyncTasksFlow is /async-tasks: the tasks running in this directory.
// Enter shows the end of the output of a task, s stops it.
func (a *app) openAsyncTasksFlow() {
	a.showAsyncTasks("")
}

func (a *app) showAsyncTasks(current string) {
	tasks, err := a.store.RunningAsyncTasks(a.dir)
	options := make([]option, len(tasks))
	for i, task := range tasks {
		detail := "running since " + relativeTime(task.StartedAt) + " · session " + shortID(task.SessionID)
		options[i] = option{label: task.ID + "  " + oneLine(task.Command), detail: detail, value: task.ID}
	}
	sel := a.openList("Async tasks · "+shortPath(a.dir), options, current, a.showAsyncOutput)
	sel.twoLines = true
	sel.empty = "No background tasks are running"
	sel.hint = "Enter output · s stop · r refresh · / search"
	sel.actions = map[rune]func(string){
		's': func(id string) { a.confirmStopAsync(id) },
		'r': func(id string) { a.showAsyncTasks(id) },
	}
	if err != nil {
		sel.err = err.Error()
	}
}

// showAsyncOutput puts the end of a task's output into the chat. It is for
// the user only; the agent does not see it.
func (a *app) showAsyncOutput(id string) error {
	task, found, err := a.store.AsyncTask(id)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("task " + id + " not found")
	}
	text, truncated, err := async.Tail(task.LogPath, asyncTailChars)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	text = strings.TrimRight(text, "\n")
	if text == "" {
		text = "(no output yet)"
	}
	head := "async task " + task.ID + " · " + task.Status + " · " + oneLine(task.Command)
	if truncated {
		head += " · last " + strconv.Itoa(asyncTailChars) + " characters"
	}
	s := a.active
	s.closeOpenEntry()
	s.appendEntry(chatEntry{kind: core.UpdateInfo, text: head + "\n" + text})
	s.scroll = 0
	return nil
}

func (a *app) confirmStopAsync(id string) {
	if id == "" {
		return
	}
	options := []option{{label: "No", value: "no"}, {label: "Yes, stop it", value: "yes"}}
	a.openList("Stop task "+id+"?", options, "no", func(answer string) error {
		if answer == "yes" {
			if err := async.Stop(id, async.StoppedByUser); err != nil {
				return err
			}
		}
		a.showAsyncTasks("")
		return nil
	})
}

func oneLine(s string) string {
	return truncate(strings.Join(strings.Fields(s), " "), 60)
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
