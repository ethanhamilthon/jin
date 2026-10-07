package session

import (
	"errors"
	"os"
	"strings"

	"jin/internal/core"
	"jin/internal/files"
	"jin/internal/prompts"
)

// Image is a picture attached to a message: Label stands in the text, Path
// is where the file is.
type Image struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

// Send queues a message the user typed. @paths become attached files,
// #prompts are expanded, images are named by their path for the model.
func (m *Manager) Send(id, text string, images []Image) error {
	return m.Do(id, func(s *Session) error {
		if !s.ready {
			return errors.New("The session is still starting")
		}
		if strings.TrimSpace(text) == "" && len(images) == 0 {
			return nil
		}
		if err := s.sendRefusal(); err != nil {
			s.emitState()
			return err
		}
		s.suggestion = ""
		home, _ := os.UserHomeDir()
		clean, paths := files.Extract(text, home, s.path)
		clean, text = attachImages(clean, text, images)
		prompt := prompts.Expand(clean, s.bodies)
		if block := files.Block(paths); block != "" {
			prompt += "\n\n" + block
		}
		s.queue(text, s.undoNote+s.todoNote()+prompt)
		s.undoNote = ""
		m.flush()
		s.emitState()
		return nil
	})
}

func (s *Session) queue(shown, prompt string) {
	request := core.Request{Prompt: prompt, Model: s.model, Effort: s.effort, Window: s.window(), NoVision: s.noVision(), Interactive: true}
	s.agent.Expect()
	s.pending = append(s.pending, request)
	s.add(Entry{Kind: core.UpdateUser, Text: shown})
	s.touch(shown)
}

// todoNote tells the model about a todo list the user edited, once.
func (s *Session) todoNote() string {
	if !s.persisted {
		return ""
	}
	edited, err := s.m.db.TakeTodosEdited(s.id)
	if err != nil || !edited {
		return ""
	}
	items, err := s.m.db.LoadTodos(s.id)
	if err != nil {
		return ""
	}
	return core.TodoEditedBlock(items)
}

// attachImages names each picture by its path for the model. A picture the
// text does not mention by its label is added at the end of both texts.
func attachImages(clean, shown string, images []Image) (string, string) {
	for _, image := range images {
		ref := strings.TrimSuffix(image.Label, "]") + ": " + image.Path + "]"
		if strings.Contains(clean, image.Label) {
			clean = strings.ReplaceAll(clean, image.Label, ref)
			continue
		}
		clean += "\n" + ref
		shown += "\n" + image.Label
	}
	return strings.TrimSpace(clean), strings.TrimSpace(shown)
}
