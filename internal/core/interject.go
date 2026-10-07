package core

import "context"

// drainPrompts takes every prompt the user queued while the model was
// working, without blocking when none are waiting. A closed channel ends
// the drain rather than spinning on its zero value. dropped counts the
// requests it took and left out: blank prompts and side requests.
func drainPrompts(prompts <-chan Request) (queued []Request, dropped int) {
	for {
		select {
		case request, open := <-prompts:
			if !open {
				return queued, dropped
			}
			if request.Kind == RequestPrompt && !request.blank() {
				queued = append(queued, request)
			} else {
				dropped++
			}
		default:
			return queued, dropped
		}
	}
}

// sendTaken reports a request that ends without a turn of its own.
func sendTaken(ctx context.Context, updates chan<- Update) bool {
	select {
	case <-ctx.Done():
		return false
	case updates <- Update{Kind: UpdateTaken}:
		return true
	}
}
