package ui

import (
	"os"
)

func (a *app) editInput(s *chatSession) error {
	if a.cfg.Editor == "" {
		a.chooseEditor(func() error { return a.editInput(s) })
		return nil
	}
	file, err := os.CreateTemp("", "jin-input-*.md")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.WriteString(draftPayload(s.input))
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return a.editFile(file.Name(), func() error {
		text, err := os.ReadFile(file.Name())
		if err != nil {
			return err
		}
		s.input = clusters(string(text))
		s.cursor, s.inputTop = len(s.input), 0
		return nil
	})
}
