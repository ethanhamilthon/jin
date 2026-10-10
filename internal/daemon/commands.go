package daemon

import (
	"context"
	"encoding/json"
	"errors"

	"jin/internal/session"
	"jin/internal/tasks"
)

type Command struct {
	ID       string          `json:"id"`
	Version  string          `json:"version"`
	Action   string          `json:"action"`
	Session  string          `json:"session,omitempty"`
	Path     string          `json:"path,omitempty"`
	Text     string          `json:"text,omitempty"`
	Shown    string          `json:"shown,omitempty"`
	Prompts  []string        `json:"prompts,omitempty"`
	Images   []session.Image `json:"images,omitempty"`
	Files    []session.File  `json:"files,omitempty"`
	Answers  []string        `json:"answers,omitempty"`
	Origin   string          `json:"origin,omitempty"`
	Question int64           `json:"question,omitempty"`
	Provider string          `json:"provider,omitempty"`
	Model    string          `json:"model,omitempty"`
	Effort   string          `json:"effort,omitempty"`
	Point    int             `json:"point,omitempty"`
}

type Result struct {
	Value json.RawMessage `json:"value,omitempty"`
	Error string          `json:"error,omitempty"`
}

func execute(ctx context.Context, manager *session.Manager, command Command) (any, error) {
	id := command.Session
	switch command.Action {
	case "providers-changed":
		return nil, manager.ProvidersChanged()
	case "settings-changed":
		manager.SettingsChanged()
		return nil, nil
	case "tasks":
		return tasks.Shared().List(""), nil
	case "task-stop":
		return nil, tasks.Shared().StopAny(command.Text)
	case "create":
		return manager.Create(command.Path)
	case "open":
		return manager.Open(id)
	case "snapshot":
		return manager.Snapshot(id)
	case "live":
		return manager.Live(), nil
	case "send-prepared":
		return nil, manager.SendPrepared(id, command.Shown, command.Text, command.Prompts, command.Images)
	case "send":
		return nil, manager.Send(id, command.Text, command.Images, command.Files)
	case "stop":
		return nil, manager.Interrupt(id)
	case "resume":
		return nil, manager.Resume(id)
	case "answer":
		return nil, manager.AnswerQuestion(id, command.Question, command.Answers)
	case "model":
		return nil, manager.SetModelProvider(id, command.Provider, command.Model, command.Effort)
	case "compact":
		return nil, manager.Compact(id)
	case "handoff":
		return nil, manager.Handoff(id, command.Origin)
	case "reload":
		return nil, manager.Reload(id)
	case "seen":
		return nil, manager.Seen(id)
	case "focus":
		return nil, manager.Focus(id)
	case "shell":
		return nil, manager.Shell(id, command.Text)
	case "fork":
		return manager.Fork(id, command.Point)
	case "points":
		return manager.Points(id)
	case "context":
		return manager.Context(id)
	case "title":
		return manager.GenerateTitle(ctx, id)
	default:
		return nil, errors.New("unknown daemon command")
	}
}
