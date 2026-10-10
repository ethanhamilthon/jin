package session

import (
	"errors"

	"jin/internal/prompts"
)

// SendPrepared preserves the TUI's distinction between tokens and pasted text.
func (m *Manager) SendPrepared(id, shown, clean string, names []string, images []Image) error {
	return m.Do(id, func(s *Session) error {
		if !s.ready {
			return errors.New("The session is still starting")
		}
		if err := s.sendRefusal(); err != nil {
			return err
		}
		clean, shown, pictures := attachImages(clean, shown, images, !s.noVision())
		prompt := prompts.ExpandNames(clean, names, s.bodies)
		s.queue(shown, prompt, pictures)
		m.flush()
		s.emitState()
		return nil
	})
}
