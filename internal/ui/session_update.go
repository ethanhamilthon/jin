package ui

import (
	"jin/internal/core"
)

func (s *chatSession) appendEntry(entry chatEntry) {
	s.continueAnswer = false
	oldRows := len(s.rows)
	if needsGap(s.lastShown(), entry.kind) {
		s.rows = append(s.rows, chatRow{})
	}
	s.history = append(s.history, entry)
	s.rows = append(s.rows, s.entryRows(entry)...)
	if s.scroll > 0 {
		s.scroll += len(s.rows) - oldRows
	}
}

func (s *chatSession) showUpdate(update core.Update) {
	if update.Kind == core.UpdateReset {
		s.dropAttempt()
		return
	}
	if update.Kind != core.UpdateAssistantDelta && update.Kind != core.UpdateReasoningDelta {
		s.closeOpenEntry()
	}
	switch update.Kind {
	case core.UpdateWorking:
		s.working = true
	case core.UpdateDone:
		s.endAttempt()
		s.working, s.changeTurn = false, 0
		s.ask = nil
	case core.UpdateUsage:
		s.applyUsage(update)
		s.persistUsage()
	case core.UpdateCompacted:
		s.applyCompacted(update)
		s.persistUsage()
		label := update.Text
		if label == "" {
			label = compactedLabel
		}
		s.appendEntry(chatEntry{kind: core.UpdateCompacted, text: label})
	case core.UpdateHistory:
		if update.Message.Role == "assistant" {
			s.endAttempt()
		}
		s.continueAnswer = update.Message.Role == "user" && core.IsTasksNote(update.Message.Content)
		s.persistMessage(update.Message)
	case core.UpdateAsk:
		s.ask = newAskState(update.Questions)
	case core.UpdateTodo:
		s.setTodos(update.Todos)
	case core.UpdateToolResult:
		s.recordChanges(update.Changes)
		s.appendEntry(chatEntry{kind: core.UpdateToolResult, tool: update.Tool, text: resultText(update.Tool, update.Text, update.Changes)})
	case core.UpdateAssistantDelta, core.UpdateReasoningDelta:
		kind := core.UpdateAssistant
		if update.Kind == core.UpdateReasoningDelta {
			kind = core.UpdateReasoning
		}
		s.appendDelta(kind, update.Text)
	default:
		s.appendEntry(chatEntry{kind: update.Kind, text: update.Text, tool: update.Tool})
	}
}

func (s *chatSession) resize(width int) {
	s.flushStream()
	if width != s.width {
		s.width = width
		s.rebuildRows(width)
	}
}
