package store

import "encoding/json"

const keyHooksDisabled = "hooks.disabled"

// parseHooksDisabled reads the names of switched-off hooks; anything
// unreadable means every hook is on.
func parseHooksDisabled(value string) []string {
	var names []string
	if value == "" || json.Unmarshal([]byte(value), &names) != nil {
		return nil
	}
	return names
}

// SaveHooksDisabled stores the names of hooks that are switched off.
func (db *DB) SaveHooksDisabled(names []string) error {
	value := ""
	if len(names) > 0 {
		data, err := json.Marshal(names)
		if err != nil {
			return err
		}
		value = string(data)
	}
	return db.setSettings(map[string]string{keyHooksDisabled: value})
}
