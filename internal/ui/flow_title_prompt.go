package ui

import (
	"errors"
	"os"
	"strconv"
)

func (a *app) openTitleAfter() {
	a.openField("After · messages before the title; 0 turns it off", strconv.Itoa(a.cfg.Title.After), false, func(text string) error {
		after, err := strconv.Atoi(text)
		if err != nil {
			return errors.New("After must be a number from 0 to 50")
		}
		t := a.cfg.Title
		t.After = after
		return a.saveTitleThen(t, "after")
	})
}

// editTitlePrompt opens the prompt in the editor, as the System prompt row
// does. An empty prompt means the built-in one.
func (a *app) editTitlePrompt() error {
	if a.cfg.Editor == "" {
		a.chooseEditor(func() error { return a.editTitlePrompt() })
		return nil
	}
	file, err := os.CreateTemp("", "jin-title-prompt-*.md")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, err = file.WriteString(a.cfg.Title.Prompt)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return a.editFile(file.Name(), func() error {
		text, err := os.ReadFile(file.Name())
		if err != nil {
			return err
		}
		t := a.cfg.Title
		t.Prompt = string(text)
		return a.saveTitleThen(t, "prompt")
	})
}
