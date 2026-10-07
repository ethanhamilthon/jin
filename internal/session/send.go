package session

import (
	"errors"
	"os"
	"strings"

	"jin/internal/core"
	"jin/internal/files"
	"jin/internal/prompts"
	"jin/internal/provider"
	"jin/internal/tools"
)

// Image is a picture attached to a message: Label stands in the text, Path
// is where the file is.
type Image struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

// Send queues a message the user typed. @paths and uploaded files become
// attached files, #prompts are expanded, images go along as pictures.
func (m *Manager) Send(id, text string, images []Image, attached []File) error {
	return m.Do(id, func(s *Session) error {
		if !s.ready {
			return errors.New("The session is still starting")
		}
		if strings.TrimSpace(text) == "" && len(images) == 0 && len(attached) == 0 {
			return nil
		}
		if err := s.sendRefusal(); err != nil {
			s.emitState()
			return err
		}
		s.suggestion = ""
		home, _ := os.UserHomeDir()
		clean, paths := files.Extract(text, home, s.path)
		paths, text = addAttached(paths, text, attached)
		clean, text, pictures := attachImages(clean, text, images, !s.noVision())
		prompt := prompts.Expand(clean, s.bodies)
		if block := files.Block(paths); block != "" {
			prompt += "\n\n" + block
		}
		s.queue(text, s.undoNote+s.todoNote()+prompt, pictures)
		s.undoNote = ""
		m.flush()
		s.emitState()
		return nil
	})
}

func (s *Session) queue(shown, prompt string, pictures []provider.Image) {
	request := core.Request{Prompt: prompt, Model: s.model, Effort: s.effort, Window: s.window(), NoVision: s.noVision(), Images: pictures, Interactive: true}
	s.agent.Expect()
	s.pending = append(s.pending, request)
	s.add(userEntry(shown, pictures))
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

// attachImages loads the pictures to send as separate message parts and takes
// their labels out of the model's text. A picture the text does not mention
// is added to the shown text. A picture that cannot be loaded stays in the
// text by its path, so the model can still try the read tool.
func attachImages(clean, shown string, images []Image, send bool) (string, string, []provider.Image) {
	var loaded []provider.Image
	for _, image := range images {
		if !strings.Contains(clean, image.Label) {
			shown += "\n" + image.Label
		} else {
			clean = strings.ReplaceAll(clean, image.Label, "")
		}
		if !send {
			continue
		}
		picture, err := tools.LoadImageFile(image.Path)
		if err != nil {
			clean += "\n" + strings.TrimSuffix(image.Label, "]") + ": " + image.Path + "]"
			continue
		}
		loaded = append(loaded, provider.Image{MimeType: picture.MimeType, Data: picture.Data})
	}
	return strings.TrimSpace(clean), strings.TrimSpace(shown), loaded
}
