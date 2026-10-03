package core

// SetContextSize restores the last reported context size before Run starts.
func (a *Agent) SetContextSize(tokens int) { a.size = max(0, tokens) }
