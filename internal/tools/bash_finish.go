package tools

import (
	"errors"
	"jin/internal/tasklog"
	"os/exec"
	"strconv"
)

// moveOn hands a running command to the background tasks, or kills it when
// that is not possible. why is the sentence that tells the agent what happened.
func moveOn(cmd *exec.Cmd, files tasklog.Files, command, dir string, background Background, canAdopt bool, finished chan error, why string) string {
	if !canAdopt {
		killProcessGroup(cmd)
		err := <-finished
		return finishBash(cmd, files, err, true)
	}
	info := background.Tasks.Adopt(background.Owner, command, dir, files, cmd, finished)
	return readLog(files.Log) + "\n[command " + why + ". It keeps running as background task " + info.ID +
		". Its result arrives as a message when it ends; do not wait for it. Use the task tool to check (limit 2000) or stop it.]"
}

// finishBash builds the result of a command that ended while we waited.
// killed is set when jin ended it because its time was up and nothing could
// adopt it.
func finishBash(cmd *exec.Cmd, files tasklog.Files, err error, killed bool) string {
	result, cut := cutOutput(files.Log)
	if cut {
	} else {
		result, _, _ = tasklog.Head(files.Log, maxBashOutput)
		tasklog.Remove(files.Log)
	}
	switch {
	case killed:
		result += "\n[command timed out]"
	case err != nil:
		result += "\n" + exitNote(err)
	case result == "":
		result = "[command completed with no output]"
	}
	rememberGroup(cmd)
	return result
}

// readLog is the start of the output of a command that moved on.
func readLog(path string) string {
	text, truncated, _ := tasklog.Head(path, maxBashOutput)
	if truncated {
		text += "\n[output truncated after " + strconv.Itoa(maxBashOutput) + " bytes; the full output keeps growing in " + path +
			": read it with the read tool or check the task]"
	}
	return text
}

// exitNote reports how a command ended without calling it a failure: a non-zero
// exit code is often an answer (grep found nothing, diff found a difference).
func exitNote(err error) string {
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() >= 0 {
		return "[exit code: " + strconv.Itoa(exit.ExitCode()) + "]"
	}
	return "[command failed: " + err.Error() + "]"
}
