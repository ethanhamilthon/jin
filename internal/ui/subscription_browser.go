package ui

import (
	"os/exec"
	"runtime"
)

func openSubscriptionBrowser(url string) error {
	name, args := "xdg-open", []string{url}
	if runtime.GOOS == "darwin" {
		name = "open"
	}
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
