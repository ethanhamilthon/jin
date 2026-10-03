package update

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Main is `jin update`: install the latest release over the running binary.
func Main(ctx context.Context, args []string, current string, out, errOut io.Writer) int {
	check := len(args) > 0 && args[0] == "--check"
	if len(args) > 1 || (len(args) == 1 && !check) {
		fmt.Fprintln(errOut, "usage: jin update [--check]")
		return 2
	}
	tag, err := Latest(ctx)
	if err != nil {
		fmt.Fprintln(errOut, "jin update:", err)
		return 1
	}
	if !Newer(tag, current) {
		fmt.Fprintf(out, "jin %s is the latest version\n", current)
		return 0
	}
	if check {
		fmt.Fprintf(out, "jin %s is available (you have %s); run jin update\n", tag, current)
		return 0
	}
	target, err := executable()
	if err != nil {
		fmt.Fprintln(errOut, "jin update:", err)
		return 1
	}
	fmt.Fprintf(out, "downloading jin %s for this system (%s)\n", tag, Archive())
	if err := Install(ctx, tag, target); err != nil {
		fmt.Fprintln(errOut, "jin update:", err)
		if os.IsPermission(err) {
			fmt.Fprintln(errOut, "try: sudo jin update")
		}
		return 1
	}
	fmt.Fprintf(out, "updated jin %s -> %s (%s)\nrestart jin to use it\n", current, tag, target)
	return 0
}

func executable() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}
