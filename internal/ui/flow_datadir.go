package ui

import (
	"errors"

	"jin/internal/async"
	"jin/internal/datadir"
)

// DataAction is a move of the data directory that has to wait until jin
// has closed its database; Run returns it and main carries it out.
type DataAction struct {
	Kind string // "reset" or "swap"
	Path string
}

// openResetFlow explains what /reset does, asks where to put the current
// data, and quits so main can move it.
func (a *app) openResetFlow() {
	current, err := datadir.Current()
	if err != nil {
		a.report(err)
		return
	}
	if err := a.dataMoveAllowed(); err != nil {
		a.report(err)
		return
	}
	options := []option{
		{label: "No, keep everything", value: "no"},
		{label: "Yes, move my data aside and start over", value: "yes"},
	}
	a.openList("Reset jin? All settings, providers, prompts, hooks, themes and sessions go away", options, "no", func(answer string) error {
		if answer != "yes" {
			return nil
		}
		a.openField("Move "+shortPath(current)+" to (a new folder you keep)", shortPath(current)+"-backup", false, func(path string) error {
			dest, err := datadir.Expand(path)
			if err != nil {
				return err
			}
			if err := datadir.CheckReset(current, dest); err != nil {
				return err
			}
			a.dataAction = &DataAction{Kind: "reset", Path: dest}
			a.quit = true
			return nil
		})
		return nil
	})
}

// openSwapFlow makes another data folder the current one; the current data
// moves to that folder's old place, so the same swap brings it back.
func (a *app) openSwapFlow() {
	current, err := datadir.Current()
	if err != nil {
		a.report(err)
		return
	}
	if err := a.dataMoveAllowed(); err != nil {
		a.report(err)
		return
	}
	a.openField("Folder to use instead of "+shortPath(current)+" (current data moves there)", "", false, func(path string) error {
		other, err := datadir.Expand(path)
		if err != nil {
			return err
		}
		if err := datadir.CheckSwap(current, other); err != nil {
			return err
		}
		a.dataAction = &DataAction{Kind: "swap", Path: other}
		a.quit = true
		return nil
	})
}

// dataMoveAllowed refuses while work is still running: a moved database
// under a running request or background task would lose its results.
func (a *app) dataMoveAllowed() error {
	if a.anyWorking() {
		return errors.New("a request is still running; stop it first")
	}
	for _, n := range a.asyncRunning {
		if n > 0 {
			return errors.New("background tasks are running; stop them in /tasks first")
		}
	}
	if err := async.StopIdle(); err != nil {
		return err
	}
	return datadir.CheckAlone()
}
