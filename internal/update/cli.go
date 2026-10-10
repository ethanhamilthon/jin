package update

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Main is `jin update`: install the latest release over the running binary.
// control stops a running daemon before the binary is replaced and starts one
// again afterwards; a zero control skips that.
func Main(ctx context.Context, args []string, current string, control Control, out, errOut io.Writer) int {
	check, force, err := updateFlags(args)
	if err != nil {
		fmt.Fprintln(errOut, "usage: jin update [--check] [--force]")
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
	if err := control.stop(ctx, force, out); err != nil {
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
	fmt.Fprintf(out, "updated jin %s -> %s (%s)\n", current, tag, target)
	if err := control.start(ctx, out); err != nil {
		fmt.Fprintln(errOut, "jin update:", err)
		fmt.Fprintln(errOut, "run jin to start the daemon again")
		return 1
	}
	return 0
}

func updateFlags(args []string) (check, force bool, err error) {
	for _, arg := range args {
		switch arg {
		case "--check":
			check = true
		case "--force":
			force = true
		default:
			return false, false, errors.New("unknown flag " + arg)
		}
	}
	return check, force, nil
}

// executable is where the running binary lives; a test replaces it.
var executable = func() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}
