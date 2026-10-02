package ui

import (
	"errors"
	"os"
	"strings"
)

func (a *app) openInputFlow() *selector {
	options := []option{
		{label: "Focus input", value: "focus"},
		{label: "Clear", value: "clear"},
		{label: "Copy", value: "copy"},
		{label: "Paste", value: "paste"},
		{label: "Edit in editor", value: "edit"},
	}
	return a.openList("Input", options, "focus", func(value string) error {
		s := a.active
		switch value {
		case "focus":
		case "clear":
			s.input, s.cursor, s.inputTop = nil, 0, 0
		case "copy":
			copySelection(a.screen, strings.Join(s.input, ""))
		case "paste":
			text, ok := pasteClipboard()
			if !ok {
				return errors.New("Could not read clipboard")
			}
			insertClusters(&s.input, &s.cursor, text)
		case "edit":
			return a.editInput(s)
		}
		return nil
	})
}

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
	_, writeErr := file.WriteString(strings.Join(s.input, ""))
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
