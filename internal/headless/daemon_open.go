package headless

import (
	"context"

	"jin/internal/session"
)

func openDaemonManaged(ctx context.Context, s *managed) (*managed, func(), error) {
	ctx, cancel := context.WithCancel(ctx)
	s.m = daemonManager{daemonClient, ctx}
	events := daemonClient.Events(ctx)
	go func() {
		for event := range events {
			if event.Type == "handoff" && event.Session == s.id {
				select {
				case s.handoff <- event.Text:
				default:
				}
			}
		}
	}()
	if _, err := daemonClient.Open(ctx, s.id); err != nil {
		cancel()
		return nil, nil, err
	}
	if err := s.wait(ctx, func(state session.State) bool { return state.Ready }); err != nil {
		cancel()
		return nil, nil, err
	}
	return s, cancel, nil
}
