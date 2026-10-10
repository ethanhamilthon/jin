package headless

// saved remembers the first failure to write the session; the run still
// finishes, but reports it and exits with 1.
func (r *runState) saved(err error) {
	if err != nil && r.saveErr == nil {
		r.saveErr = err
	}
}
