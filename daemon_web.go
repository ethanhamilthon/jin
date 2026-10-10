package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"jin/internal/daemon"
	"jin/internal/datadir"
	"jin/internal/web"
)

func daemonWeb(args []string, dir string) (int, error) {
	opt, err := web.ParseArgs(args)
	if err != nil {
		return 2, err
	}
	if opt.Help {
		fmt.Fprint(os.Stdout, web.Usage)
		return 0, nil
	}
	if opt.Cwd != "" {
		dir = opt.Cwd
	}
	root, err := datadir.Current()
	if err != nil {
		return 1, err
	}
	client, err := daemon.Ensure(context.Background(), root, version)
	if err != nil {
		return 1, err
	}
	link, err := client.Web(context.Background(), dir, opt)
	if err != nil {
		return 1, err
	}
	fmt.Println("jin web is running at", link)
	if !opt.NoOpen {
		name, args := "", []string{link}
		switch runtime.GOOS {
		case "darwin":
			name = "open"
		case "linux":
			name = "xdg-open"
		}
		if name != "" {
			_ = exec.Command(name, args...).Run()
		}
	}
	return 0, nil
}
