package ui

import (
	"strconv"
	"strings"

	"jin/internal/core"
)

func (a *app) finishReload(s *chatSession, ev renderEvent) {
	if s.render == nil || !s.render.reload {
		return
	}
	s.render.cancel()
	s.render = nil
	if a.store != nil {
		a.retryAsync(s.id)
	}
	switch {
	case ev.cancelled:
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Prompt reload cancelled"})
	case ev.err != "":
		s.appendEntry(chatEntry{kind: core.UpdateError, text: "Prompt reload failed: " + ev.err})
	case ev.out == nil:
		s.appendEntry(chatEntry{kind: core.UpdateError, text: "Prompt reload failed: no render output"})
	default:
		s.promptBodies = ev.out.Prompts
		s.customSystem = ev.out.Custom
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Reloaded system, compact, and handoff prompts: " + reloadCounts(ev.hookCount, len(ev.out.Prompts))})
		if len(ev.out.Warnings) > 0 {
			s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Prompt reload warnings: " + strings.Join(ev.out.Warnings, "; ")})
		}
	}
}

func reloadCounts(hooks, prompts int) string {
	return strings.Join([]string{countLabel(hooks, "hook"), countLabel(prompts, "#prompt")}, ", ")
}

func countLabel(count int, name string) string {
	if count == 1 {
		return "1 " + name
	}
	return strconv.Itoa(count) + " " + name + "s"
}
