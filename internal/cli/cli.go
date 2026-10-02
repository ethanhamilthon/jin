// Package cli decides what `jin <args>` means. Only a bare `jin` opens the
// TUI; everything else is a headless command, or an honest error.
package cli

import (
	"fmt"
	"io"

	"jin/internal/headless"
)

// Kind is the command a command line selects.
type Kind int

const (
	TUI Kind = iota
	Headless
	Async
	Daemon
	Help
	Version
	Unknown
)

// ExitUsage is the exit code of an unknown command.
const ExitUsage = 2

// Classify maps the arguments (without the program name) to a command.
func Classify(args []string) Kind {
	if len(args) == 0 {
		return TUI
	}
	switch args[0] {
	case "help", "--help", "-h":
		return Help
	case "version", "--version", "-v":
		return Version
	case "async":
		return Async
	case "daemon":
		return Daemon
	}
	if headless.Handles(args) {
		return Headless
	}
	return Unknown
}

// NeedsDB reports whether the command opens the database; help, version and
// unknown commands never touch it.
func NeedsDB(kind Kind) bool {
	return kind != Help && kind != Version && kind != Unknown
}

const usage = `jin: a minimal terminal coding agent

Usage:
  jin                          open the TUI
  jin -p [flags] [prompt...]   run one request without the TUI (see docs/headless.md)
  jin models [--all]           list the models you use, with price and context window
  jin refresh-models           refresh the cached model list
  jin async run "<cmd>" --session <id>
                               run a command in the background; its result
                               comes back to the session as a message
  jin async check --id <task> [--limit <n>]
                               status and the last n characters of the output
  jin async input --id <task> "<text>"
                               write a line to the stdin of a task
  jin async stop --id <task>   stop a task
  jin --version                print the version
  jin --help                   print this help
`

// PrintHelp writes the usage text.
func PrintHelp(w io.Writer) {
	fmt.Fprint(w, usage)
}

// PrintVersion writes the version line.
func PrintVersion(w io.Writer, version string) {
	fmt.Fprintln(w, "jin", version)
}

// PrintUnknown reports a command jin does not have and points to --help.
func PrintUnknown(w io.Writer, args []string) {
	fmt.Fprintf(w, "jin: unknown command %q\nRun 'jin --help' for usage.\n", args[0])
}
