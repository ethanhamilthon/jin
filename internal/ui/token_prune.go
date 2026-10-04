package ui

// pruneTokens forgets the tokens that no open draft holds any more, so sent
// or deleted pastes are not kept for the life of the process.
func (a *app) pruneTokens() {
	held := map[string]bool{}
	for _, s := range a.sessions {
		for _, cluster := range s.input {
			held[cluster] = true
		}
	}
	for _, cluster := range a.active.input {
		held[cluster] = true
	}
	tokens.Lock()
	defer tokens.Unlock()
	for mark := range tokens.byMark {
		if !held[mark] {
			delete(tokens.byMark, mark)
		}
	}
}
