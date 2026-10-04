package tools

import (
	"os"
	"os/exec"
)

var nonInteractiveEnv = []string{
	"GIT_TERMINAL_PROMPT=0",
	"GIT_EDITOR=true",
	"GIT_PAGER=cat",
	"PAGER=cat",
	"DEBIAN_FRONTEND=noninteractive",
}

// setGroup detaches the command from the terminal of jin and makes programs
// that would ask questions or open an editor fail or carry on instead.
func setGroup(cmd *exec.Cmd) {
	setSession(cmd)
	cmd.Env = append(os.Environ(), nonInteractiveEnv...)
}

// Isolate applies the setup of the bash tool to a command started elsewhere,
// such as the user's ! shell.
func Isolate(cmd *exec.Cmd) { setGroup(cmd) }

// KillGroup ends the whole process group of a started command.
func KillGroup(cmd *exec.Cmd) { killProcessGroup(cmd) }
