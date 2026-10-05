// Package tasks runs background commands for the agents of this jin process.
// A task lives as long as the process: jin watches it, keeps its log bounded,
// and announces its end as an Event. Nothing outlives jin.
package tasks

import (
	"os"
	"os/exec"
	"time"
)

// Status of a task.
const (
	Running = "running"
	Done    = "done"
	Failed  = "failed"
	Stopped = "stopped"
)

// Info is what callers may see of a task.
type Info struct {
	ID, Owner, Command, Dir, Log, Status string
	Exit                                 int
	Started                              time.Time
	Stdin                                bool
}

type task struct {
	info    Info
	cmd     *exec.Cmd
	stdin   *os.File
	writing chan struct{}
	stopBy  string
	ended   chan struct{}
}

func (t *task) closeStdin() {
	if t.stdin != nil {
		_ = t.stdin.Close()
	}
}
