//go:build unix

package async

import (
	"fmt"
	"jin/internal/store"
	"jin/internal/tasklog"
	"os"
	"strconv"
	"strings"
	"time"
)

// endAdopted records how an adopted task ended. The exit code comes from the
// file jin writes; when jin was closed first there is none, and the task is
// reported as ended with an unknown code.
func (d *daemon) endAdopted(t store.AsyncTask) {
	defer d.forget(t.ID)
	code, known := readExit(t.ExitPath)
	for deadline := time.Now().Add(exitGrace); !known && time.Now().Before(deadline); {
		time.Sleep(100 * time.Millisecond)
		code, known = readExit(t.ExitPath)
	}
	status, exit := store.AsyncDone, code
	switch {
	case !known:
		status, exit = store.AsyncEnded, -1
	case code != 0:
		status = store.AsyncFailed
	}
	stored := code
	if !known {
		stored = 0
	}
	won, err := d.db.FinishAsyncTask(t.ID, status, stored)
	if err != nil || !won {
		return
	}
	output, truncated, _ := Tail(t.LogPath, eventTail)
	if truncated {
		output = fmt.Sprintf("[output cut: last %d characters; run `jin async check --id %s` for more]\n%s", eventTail, t.ID, output)
	}
	if !known {
		output = "[the exit code is unknown: jin was closed while the command ran]\n" + output
	}
	_ = d.db.AddAsyncEvent(t.SessionID, t.Path, ResultText(t.ID, status, exit, output))
	tasklog.Remove(t.ExitPath)
}

func readExit(path string) (int, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	code, err := strconv.Atoi(strings.TrimSpace(string(data)))
	return code, err == nil
}
