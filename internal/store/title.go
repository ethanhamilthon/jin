package store

import (
	"errors"
	"strconv"
	"strings"
)

const (
	keyTitleProvider = "title.provider"
	keyTitleModel    = "title.model"
	keyTitleEffort   = "title.effort"
	keyTitlePrompt   = "title.prompt"
	keyTitleAfter    = "title.after"
)

const (
	DefaultTitleAfter  = 4
	maxTitleAfter      = 50
	DefaultTitlePrompt = "Write a short title for this conversation: 2 to 6 words that name its subject, " +
		"in the language the user writes in. Reply with the title only: no quotes, no full stop, no explanation."
)

// TitleSettings configures the session titles. An empty Provider and Model
// mean the model of the session; an empty Prompt means DefaultTitlePrompt;
// After is the message count that titles a session, 0 turns it off.
type TitleSettings struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Effort   string `json:"effort"`
	Prompt   string `json:"prompt"`
	After    int    `json:"after"`
}

func parseTitle(values map[string]string) TitleSettings {
	t := TitleSettings{Provider: values[keyTitleProvider], Model: values[keyTitleModel], Effort: values[keyTitleEffort],
		Prompt: values[keyTitlePrompt], After: DefaultTitleAfter}
	if n, err := strconv.Atoi(values[keyTitleAfter]); err == nil && n >= 0 {
		t.After = min(n, maxTitleAfter)
	}
	if strings.TrimSpace(t.Prompt) == "" {
		t.Prompt = DefaultTitlePrompt
	}
	return t
}

// SaveTitle stores the title settings. The default prompt is stored as empty,
// so a later change of the default reaches it.
func (db *DB) SaveTitle(t TitleSettings) error {
	if t.Model != "" && t.Provider == "" {
		return errors.New("the title model needs its provider")
	}
	if t.After < 0 || t.After > maxTitleAfter {
		return errors.New("title after must be 0 to 50")
	}
	prompt := strings.TrimSpace(t.Prompt)
	if prompt == DefaultTitlePrompt {
		prompt = ""
	}
	return db.setSettings(map[string]string{keyTitleProvider: t.Provider, keyTitleModel: t.Model, keyTitleEffort: t.Effort,
		keyTitlePrompt: prompt, keyTitleAfter: strconv.Itoa(t.After)})
}

// SetTitle renames a stored session.
func (db *DB) SetTitle(id, title string) error {
	return retryBusy(func() error {
		_, err := db.sql.Exec(`UPDATE sessions SET title = ? WHERE id = ?`, title, id)
		return err
	})
}
