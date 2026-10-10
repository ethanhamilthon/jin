package main

import (
	"errors"
	"fmt"
	"os"

	"jin/internal/cliproxy"
	"jin/internal/datadir"
	"jin/internal/store"
	"jin/internal/tasks"
	"jin/internal/tools"
)

// moveData carries out /reset or /swap-config once the TUI is gone: the
// database must be closed before its folder moves.
func moveData(db *store.DB, action *datadir.Action) (int, error) {
	if len(tasks.Shared().Running("")) > 0 {
		return 1, errors.New("background tasks are still running; nothing was moved")
	}
	tools.KillBackground()
	cliproxy.CloseAll()
	if err := datadir.Exclusive(); err != nil {
		return 1, fmt.Errorf("%w; nothing was moved", err)
	}
	if err := db.Close(); err != nil {
		return 1, err
	}
	current, err := datadir.Current()
	if err != nil {
		return 1, err
	}
	switch action.Kind {
	case "reset":
		if err := datadir.Reset(current, action.Path); err != nil {
			return 1, err
		}
		fmt.Fprintf(os.Stdout, "jin data moved to %s\nrun jin to start from scratch; to go back, use /swap-config with that folder\n", action.Path)
	case "swap":
		if err := datadir.Swap(current, action.Path); err != nil {
			return 1, err
		}
		fmt.Fprintf(os.Stdout, "jin now uses the data from %s\nthe previous data is in %s; run jin to continue\n", action.Path, action.Path)
	}
	return 0, nil
}
