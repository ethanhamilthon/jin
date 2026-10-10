package ui

import (
	"jin/internal/daemon"
	"jin/internal/tasks"
)

func (a *app) taskList() []tasks.Info {
	if a.backend == nil {
		return tasks.Shared().List("")
	}
	var list []tasks.Info
	if err := a.backend.Command(a.ctx, daemon.Command{Action: "tasks"}, &list); err != nil {
		a.backendError(err)
	}
	return list
}

func (a *app) stopTask(id string) error {
	if a.backend == nil {
		return tasks.Shared().StopAny(id)
	}
	return a.backend.Command(a.ctx, daemon.Command{Action: "task-stop", Text: id}, nil)
}
