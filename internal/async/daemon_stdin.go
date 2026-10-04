//go:build unix

package async

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// inputTimeout bounds one `jin async input`: waiting for an earlier input of
// the same task and writing to a child that does not read.
var inputTimeout = 5 * time.Second

type task struct {
	stdin   *os.File
	writing chan struct{}
}

func newTask(stdin *os.File) *task {
	return &task{stdin: stdin, writing: make(chan struct{}, 1)}
}

func (t *task) closeStdin() {
	if t != nil && t.stdin != nil {
		_ = t.stdin.Close()
	}
}

func (d *daemon) input(req request) error {
	d.mu.Lock()
	t := d.tasks[req.ID]
	d.mu.Unlock()
	if t == nil {
		return fmt.Errorf("task %s is not running", req.ID)
	}
	if t.stdin == nil {
		return fmt.Errorf("task %s has no stdin you can write to; only tasks started with `jin async run --stdin` have one", req.ID)
	}
	text := req.Text
	if !req.NoNewline {
		text += "\n"
	}
	select {
	case t.writing <- struct{}{}:
		defer func() { <-t.writing }()
	case <-time.After(inputTimeout):
		return fmt.Errorf("task %s is still busy with an earlier input", req.ID)
	}
	_ = t.stdin.SetWriteDeadline(time.Now().Add(inputTimeout))
	_, err := io.WriteString(t.stdin, text)
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return fmt.Errorf("task %s does not read its stdin; the input was not fully written", req.ID)
	}
	return err
}
