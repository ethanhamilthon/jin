package core

// drainPrompts takes every prompt the user queued while the model was
// working, without blocking when none are waiting. A closed channel ends
// the drain rather than spinning on its zero value.
func drainPrompts(prompts <-chan Request) []Request {
	var queued []Request
	for {
		select {
		case request, open := <-prompts:
			if !open {
				return queued
			}
			if request.Kind == RequestPrompt && !request.blank() {
				queued = append(queued, request)
			}
		default:
			return queued
		}
	}
}
