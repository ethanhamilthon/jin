package hooks

import (
	"context"
	"fmt"
	"io"
	"strings"
)

const cliUsage = `usage:
  jin hooks list                               list global and project hooks
  jin hooks add <url|path> [--name n] [--project]
                                               copy a markdown hook into ~/.jin/hooks,
                                               or into ./.jin/hooks with --project`

// Main runs `jin hooks`. dir is the working directory.
func Main(ctx context.Context, args []string, dir string, out, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(errOut, cliUsage)
		return 2
	}
	switch args[0] {
	case "list":
		return list(dir, out)
	case "add":
		return add(ctx, args[1:], dir, out, errOut)
	}
	fmt.Fprintln(errOut, cliUsage)
	return 2
}

func list(dir string, out io.Writer) int {
	global, _ := List()
	for _, name := range global {
		fmt.Fprintln(out, name)
	}
	project, _ := ListProject(dir)
	for _, name := range project {
		fmt.Fprintln(out, name+"\t(project)")
	}
	return 0
}

func add(ctx context.Context, args []string, dir string, out, errOut io.Writer) int {
	var source, name, projectDir string
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; {
		case arg == "--project":
			projectDir = dir
		case arg == "--name" && i+1 < len(args):
			name = args[i+1]
			i++
		case strings.HasPrefix(arg, "--name="):
			name = strings.TrimPrefix(arg, "--name=")
		case strings.HasPrefix(arg, "-") || source != "":
			fmt.Fprintln(errOut, cliUsage)
			return 2
		default:
			source = arg
		}
	}
	if source == "" {
		fmt.Fprintln(errOut, cliUsage)
		return 2
	}
	path, err := Add(ctx, source, name, projectDir)
	if err != nil {
		fmt.Fprintln(errOut, "jin:", err)
		return 1
	}
	fmt.Fprintln(out, "Added", path)
	fmt.Fprintln(out, "Read it before you use it: its {{commands}} run when a session starts. It applies to new sessions.")
	return 0
}
