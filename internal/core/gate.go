package core

import "jin/internal/provider"

// SetRequestGate sets a check that runs right before every model request of
// a turn. It gets the usage of the response that just completed (zero before
// the first request); an error ends the turn without sending the request.
// Call it before Run.
func (a *Agent) SetRequestGate(gate func(last provider.Usage) error) { a.gate = gate }

func (a *Agent) passGate(last provider.Usage) error {
	if a.gate == nil {
		return nil
	}
	return a.gate(last)
}
