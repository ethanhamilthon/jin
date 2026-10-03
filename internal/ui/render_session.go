package ui

import (
	"context"
	"slices"

	"jin/internal/prompts"
	"jin/internal/provider"
	"jin/internal/startup"
)

// rendering is the start-up of one session. Its commands run in the
// background and its agent starts when they are done. Until then the session
// is on screen but its input is closed.
type rendering struct {
	cancel context.CancelFunc
	// loading lists the #prompts whose commands still run.
	loading []string
}

// renderEvent is what a background render tells the main loop: a prompt that
// is done, or, with out set, the end of the whole render.
type renderEvent struct {
	session string
	prompt  string
	out     *startup.Output
	// cancelled is true when the user stopped the render.
	cancelled bool
}

// begin starts the commands of a session in the background. The agent is
// started by finishRender, with the texts that come out.
func (a *app) beginRender(s *chatSession, ctx context.Context, names []string, messages []provider.Message) {
	in := startup.Input{
		Dir: a.dir, SessionID: s.id, ToolNames: names,
		HooksDisabled: a.cfg.HooksDisabled, PromptsDisabled: a.cfg.PromptsDisabled, WithPrompts: true,
		ProjectHooks: a.projectHooksTrusted(),
	}
	renderCtx, cancel := context.WithCancel(ctx)
	s.render = &rendering{cancel: cancel}
	s.initial = messages
	infos, _ := prompts.ListInfo()
	for _, info := range infos {
		if !slices.Contains(a.cfg.PromptsDisabled, info.Name) {
			s.render.loading = append(s.render.loading, info.Name)
		}
	}
	s.syncLoading()
	id := s.id
	go func() {
		out := startup.Render(renderCtx, in, func(name string) {
			select {
			case a.rendered <- renderEvent{session: id, prompt: name}:
			case <-ctx.Done():
			}
		})
		ev := renderEvent{session: id, out: &out, cancelled: renderCtx.Err() != nil && ctx.Err() == nil}
		select {
		case a.rendered <- ev:
		case <-ctx.Done():
		}
	}()
}

// receiveRender applies one message of a background render.
func (a *app) receiveRender(ev renderEvent) {
	s, ok := a.sessions[ev.session]
	if !ok || s.render == nil {
		return
	}
	if ev.out == nil {
		s.render.loading = slices.DeleteFunc(s.render.loading, func(n string) bool { return n == ev.prompt })
		s.syncLoading()
		return
	}
	a.finishRender(s, *ev.out, ev.cancelled)
}
