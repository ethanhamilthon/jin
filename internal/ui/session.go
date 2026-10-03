package ui

import (
	"context"
	"fmt"
	"strings"

	"jin/internal/core"
	"jin/internal/pricing"
	"jin/internal/prompts"
	"jin/internal/provider"
	"jin/internal/store"
	"jin/internal/todo"
)

// chatSession is one conversation with its own backend loop. It keeps running
// in the background while another session is focused.
type chatSession struct {
	id           string
	path         string
	store        *store.DB
	provider     string
	client       *provider.Client
	persisted    bool
	agent        *core.Agent
	prompts      chan<- core.Request
	stop         context.CancelFunc
	pending      []core.Request
	history      []chatEntry
	rows         []chatRow
	input        []string
	cursor       int
	inputTop     int
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
	fold         foldMode
	todos        []todo.Item
	// changeTurn numbers the file changes of the running turn for /undo;
	// undoNote tells the model about an undo with the next message.
	changeTurn int
	undoNote   string
	todoTop      int
	ask          *askState
	bash         *bashState
	// ready is false while the session starts: its commands run in the
	// background, its agent is not running and its input is closed.
	ready  bool
	render *rendering
	// promptBodies are the #prompts, with their commands run, as of the start.
	promptBodies map[string]string
	// customSystem is true when ~/.jin/system-prompt.md was used.
	customSystem bool
	// initial, runCtx and updatesOut are what the agent needs to start.
	initial    []provider.Message
	runCtx     context.Context
	updatesOut chan core.Update
	requests   <-chan core.Request
}

// viewport is where the timeline was last drawn, for mouse hit-testing.
type viewport struct {
	first, height int
}

func (s *chatSession) appendEntry(entry chatEntry) {
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
	if update.Kind != core.UpdateAssistantDelta && update.Kind != core.UpdateReasoningDelta {
		s.closeOpenEntry()
	}
	switch update.Kind {
	case core.UpdateWorking:
		s.working = true
	case core.UpdateDone:
		s.working, s.changeTurn = false, 0
		s.ask = nil
	case core.UpdateUsage:
		s.applyUsage(update)
		s.persistUsage()
	case core.UpdateCompacted:
		s.applyCompacted(update)
		s.persistUsage()
		s.appendEntry(chatEntry{kind: core.UpdateCompacted, text: compactedLabel})
	case core.UpdateHistory:
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
		if delta := s.appendDelta(kind, update.Text); s.scroll > 0 {
			s.scroll += delta
		}
	default:
		s.appendEntry(chatEntry{kind: update.Kind, text: update.Text, tool: update.Tool})
	}
}

func (s *chatSession) send(text string) { s.sendFiles(text, text, "") }

// sendFiles shows text in the chat and sends clean plus the attached-files
// block to the model.
func (s *chatSession) sendFiles(text, clean, block string) {
	s.closeOpenEntry()
	prompt := prompts.Expand(clean, s.promptBodies)
	if block != "" {
		prompt += "\n\n" + block
	}
	note := s.undoNote + s.todoNote()
	s.undoNote = ""
	request := core.Request{Prompt: note + prompt, Model: s.model, Effort: s.effort, Window: s.window(), NoVision: s.noVision()}
	if strings.TrimSpace(request.Prompt) != "" {
		// A command that runs now moves to the background, not in your way.
		request.Interactive = true
		s.agent.Expect()
	}
	s.pending = append(s.pending, request)
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

// todoNote tells the model about a todo list the user edited, once.
func (s *chatSession) todoNote() string {
	if !s.persisted {
		return ""
	}
	edited, err := s.store.TakeTodosEdited(s.id)
	if err != nil || !edited {
		return ""
	}
	items, err := s.store.LoadTodos(s.id)
	if err != nil {
		return ""
	}
	return core.TodoEditedBlock(items)
}

// setTodos shows a new list. A finished list leaves the pin and goes into
// the timeline once.
func (s *chatSession) setTodos(items []todo.Item) {
	s.todos = items
	if todo.AllDone(items) {
		s.appendEntry(chatEntry{kind: core.UpdateTodo, text: todoEntryText(items)})
	}
}

func todoEntryText(items []todo.Item) string {
	done, total := todo.Counts(items)
	return fmt.Sprintf("Todo %d/%d\n%s", done, total, todo.Text(items))
}

// pinnedTodos is the list shown above the input, if any.
func (s *chatSession) pinnedTodos() []todo.Item {
	if len(s.todos) == 0 || todo.AllDone(s.todos) {
		return nil
	}
	return s.todos
}
