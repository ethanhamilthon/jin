//go:build desktop

package main

import (
	"fmt"
	"os"

	"jin/internal/daemon"
	"jin/internal/desktop"
)

// version must match the daemon the desktop app starts. Release builds set it with -X.
var version = "v0.14.7"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--internal-daemon" {
		if err := daemon.Serve(version); err != nil {
			fmt.Fprintln(os.Stderr, "jin:", err)
			os.Exit(1)
		}
		return
	}
	if err := desktop.Run(version); err != nil {
		fmt.Fprintln(os.Stderr, "jin desktop:", err)
		os.Exit(1)
	}
}
