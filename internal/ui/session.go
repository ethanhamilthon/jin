package ui

import (
	"context"

	"jin/internal/core"
	"jin/internal/pricing"
	"jin/internal/prompts"
	"jin/internal/store"
)

// chatSession is one conversation with its own backend loop. It keeps running
// in the background while another session is focused.
type chatSession struct {
	id           string
	path         string
	store        *store.DB
	persisted    bool
	agent        *core.Agent
	prompts      chan<- core.Request
	stop         context.CancelFunc
	pending      []core.Request
	history      []chatEntry
	rows         []chatRow
	input        []string
	cursor       int
	model        string
	effort       string
	title        string
	usage        store.Usage
	cache        cacheRate
	pricing      pricing.Table
	scroll       int
	width        int
	working      bool
	unread       bool
	openKind     core.UpdateKind
	openRowStart int
	view         viewport
	selection    textSelection
}

// viewport is where the timeline was last drawn, for mouse hit-testing.
type viewport struct {
	first, height int
}

func (s *chatSession) appendEntry(entry chatEntry) {
	oldRows := len(s.rows)
	if needsGap(s.lastEntry(), entry.kind) {
		s.rows = append(s.rows, chatRow{})
	}
	s.history = append(s.history, entry)
	s.rows = append(s.rows, entryRows(entry, s.width)...)
	if s.scroll > 0 {
		s.scroll += len(s.rows) - oldRows
	}
}

func (s *chatSession) showUpdate(update core.Update) {
	if update.Kind != core.UpdateAssistantDelta && update.Kind != core.UpdateReasoningDelta {
		s.closeOpenEntry()
	}
	switch update.Kind {
	case core.UpdateWorking:
		s.working = true
	case core.UpdateDone:
		s.working = false
	case core.UpdateUsage:
		s.applyUsage(update)
		s.persistUsage()
	case core.UpdateCompacted:
		s.applyCompacted(update)
		s.persistUsage()
		s.appendEntry(chatEntry{kind: core.UpdateCompacted, text: compactedLabel})
	case core.UpdateHistory:
		s.persistMessage(update.Message)
	case core.UpdateAssistantDelta, core.UpdateReasoningDelta:
		kind := core.UpdateAssistant
		if update.Kind == core.UpdateReasoningDelta {
			kind = core.UpdateReasoning
		}
		if delta := s.appendDelta(kind, update.Text); s.scroll > 0 {
			s.scroll += delta
		}
	default:
		s.appendEntry(chatEntry{kind: update.Kind, text: update.Text, tool: update.Tool})
	}
}

func (s *chatSession) send(text string) {
	s.closeOpenEntry()
	s.pending = append(s.pending, core.Request{Prompt: prompts.Expand(text), Model: s.model, Effort: s.effort, Window: s.window()})
	s.appendEntry(chatEntry{kind: core.UpdateUser, text: text})
	s.scroll = 0
	s.touch(text)
}

func (s *chatSession) resize(width int) {
	if width != s.width {
		s.width = width
		s.rebuildRows(width)
	}
}
