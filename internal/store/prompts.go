package store

import "encoding/json"

const keyPromptsDisabled = "prompts.disabled"

// parsePromptsDisabled reads the names of switched-off prompts; anything
// unreadable means every prompt is on.
func parsePromptsDisabled(value string) []string {
	var names []string
	if value == "" || json.Unmarshal([]byte(value), &names) != nil {
		return nil
	}
	return names
}

// SavePromptsDisabled stores the names of prompts that are switched off.
func (db *DB) SavePromptsDisabled(names []string) error {
	value := ""
	if len(names) > 0 {
		data, err := json.Marshal(names)
		if err != nil {
			return err
		}
		value = string(data)
	}
	return db.setSettings(map[string]string{keyPromptsDisabled: value})
}
