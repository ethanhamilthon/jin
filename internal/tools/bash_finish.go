package tools

import (
	"jin/internal/tasklog"
	"os"
	"os/exec"
	"strconv"
)

// moveOn hands a running command to the daemon, or kills it when that is not
// possible. why is the sentence that tells the agent what happened.
func moveOn(cmd *exec.Cmd, files tasklog.Files, command string, background Background, canAdopt bool, finished chan error, why string) string {
	if !canAdopt {
		killProcessGroup(cmd)
		err := <-finished
		return finishBash(cmd, files, err, true)
	}
	// The exit code is written by jin when the process ends, because the
	// daemon is not its parent and cannot wait for it.
	go func() {
		err := <-finished
		writeExit(files.Exit, cmd, err)
	}()
	id, err := background.Adopt(Adoption{PID: cmd.Process.Pid, PGID: cmd.Process.Pid, Command: command, Log: files.Log, Exit: files.Exit})
	if err != nil {
		killProcessGroup(cmd)
		tasklog.Remove(files.Log, files.Exit)
		return readLog(files.Log) + "\n[command " + why + "; it could not move to the background (" + err.Error() + ") and was stopped]"
	}
	return readLog(files.Log) + "\n[command " + why + ". It keeps running as background task " + id +
		". Its result will arrive as a message when it ends. Use `jin async check --id " + id +
		" --limit 2000` to look at it and `jin async stop --id " + id + "` to stop it.]"
}

// finishBash builds the result of a command that ended while we waited.
// killed is set when jin ended it because its time was up and nothing could
// adopt it.
func finishBash(cmd *exec.Cmd, files tasklog.Files, err error, killed bool) string {
	result, cut := cutOutput(files.Log)
	if cut {
		tasklog.Remove(files.Exit)
	} else {
		result, _, _ = tasklog.Head(files.Log, maxBashOutput)
		tasklog.Remove(files.Log, files.Exit)
	}
	switch {
	case killed:
		result += "\n[command timed out]"
	case err != nil:
		result += "\n[command failed: " + err.Error() + "]"
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
			": read it with the read tool or jin async check]"
	}
	return text
}

// writeExit records how a command ended, for the daemon. It writes to a
// temporary name first, so the daemon never reads a half-written file.
func writeExit(path string, cmd *exec.Cmd, err error) {
	code := 0
	if cmd.ProcessState != nil {
		code = cmd.ProcessState.ExitCode()
		if code < 0 {
			code = signalCode(cmd)
		}
	} else if err != nil {
		code = 1
	}
	tmp := path + ".tmp"
	if os.WriteFile(tmp, []byte(strconv.Itoa(code)+"\n"), 0o600) == nil {
		_ = os.Rename(tmp, path)
	}
}
