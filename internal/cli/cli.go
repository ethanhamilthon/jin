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
	Export
	Hooks
	Help
	Version
	Update
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
	case "export":
		return Export
	case "hooks":
		return Hooks
	case "update":
		return Update
	}
	if headless.Handles(args) {
		return Headless
	}
	return Unknown
}

// NeedsDB reports whether the command opens the database; help, version and
// unknown commands never touch it.
func NeedsDB(kind Kind) bool {
	return kind != Help && kind != Version && kind != Update && kind != Unknown
}

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
