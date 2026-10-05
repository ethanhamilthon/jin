package tools

import (
	"fmt"
	"strings"
	"time"

	"jin/internal/tasks"
)

func describe(info tasks.Info) string {
	line := fmt.Sprintf("task %s %s after %s: %s", info.ID, info.Status, time.Since(info.Started).Round(time.Second), info.Command)
	if info.Exit >= 0 {
		line += fmt.Sprintf(" (exit %d)", info.Exit)
	}
	return line
}

func listTasks(list []tasks.Info) string {
	if len(list) == 0 {
		return "no tasks in this session"
	}
	lines := make([]string, len(list))
	for i, info := range list {
		lines[i] = describe(info)
	}
	return strings.Join(lines, "\n")
}
