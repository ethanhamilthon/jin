package ui

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"jin/internal/core"
	"jin/internal/tasklog"
	"jin/internal/tasks"
)

const taskTailChars = 4000

// openTasksFlow is /tasks: the background tasks of this jin.
// Enter shows the end of the output of a task, s stops it.
func (a *app) openTasksFlow() {
	a.showTasks("")
}

func (a *app) showTasks(current string) {
	list := a.taskList()
	options := make([]option, len(list))
	for i, task := range list {
		detail := task.Status + " · started " + relativeTime(task.Started) + " · session " + shortID(task.Owner)
		options[i] = option{label: task.ID + "  " + oneLine(task.Command), detail: detail, value: task.ID}
	}
	sel := a.openList("Tasks", options, current, a.showTaskOutput)
	sel.twoLines = true
	sel.empty = "No background tasks"
	sel.hint = "Enter output · s stop · r refresh · / search"
	sel.actions = map[rune]func(string){
		's': func(id string) { a.confirmStopTask(id) },
		'r': func(id string) { a.showTasks(id) },
	}
}

// showTaskOutput puts the end of a task's output into the chat. It is for
// the user only; the agent does not see it.
func (a *app) showTaskOutput(id string) error {
	var task tasks.Info
	for _, info := range a.taskList() {
		if info.ID == id {
			task = info
		}
	}
	if task.ID == "" {
		return errors.New("task " + id + " not found")
	}
	text, truncated, err := tasklog.Tail(task.Log, taskTailChars)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	text = strings.TrimRight(text, "\n")
	if text == "" {
		text = "(no output yet)"
	}
	head := "task " + task.ID + " · " + task.Status + " · " + oneLine(task.Command)
	if truncated {
		head += " · last " + strconv.Itoa(taskTailChars) + " characters"
	}
	s := a.active
	s.closeOpenEntry()
	s.appendEntry(chatEntry{kind: core.UpdateInfo, text: head + "\n" + text})
	s.scroll = 0
	return nil
}

func (a *app) confirmStopTask(id string) {
	if id == "" {
		return
	}
	options := []option{{label: "No", value: "no"}, {label: "Yes, stop it", value: "yes"}}
	a.openList("Stop task "+id+"?", options, "no", func(answer string) error {
		if answer == "yes" {
			if err := a.stopTask(id); err != nil {
				return err
			}
			time.Sleep(50 * time.Millisecond)
		}
		a.showTasks("")
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
