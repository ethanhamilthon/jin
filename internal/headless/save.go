package headless

import "jin/internal/tools"

// saved remembers the first failure to write the session; the run still
// finishes, but reports it and exits with 1.
func (r *runState) saved(err error) {
	if err != nil && r.saveErr == nil {
		r.saveErr = err
	}
}

// recordChanges saves the files a tool call changed, all under one turn, so
// /undo in the TUI can revert them.
func (r *runState) recordChanges(changes []tools.Change) {
	if !r.save || len(changes) == 0 {
		return
	}
	if r.changeTurn == 0 {
		turn, err := r.db.NextTurn(r.id)
		if err != nil {
			r.saved(err)
			return
		}
		r.changeTurn = turn
	}
	for _, change := range changes {
		r.saved(r.db.SaveChange(r.id, r.changeTurn, change))
	}
}
