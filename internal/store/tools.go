package store

import "encoding/json"

const keyToolsDisabled = "tools.disabled"

// parseToolsDisabled reads the names of switched-off tools; anything
// unreadable means every tool is on.
func parseToolsDisabled(value string) []string {
	var names []string
	if value == "" || json.Unmarshal([]byte(value), &names) != nil {
		return nil
	}
	return names
}

// SaveToolsDisabled stores the names of tools that are switched off.
func (db *DB) SaveToolsDisabled(names []string) error {
	value := ""
	if len(names) > 0 {
		data, err := json.Marshal(names)
		if err != nil {
			return err
		}
		value = string(data)
	}
	return db.setSettings(map[string]string{keyToolsDisabled: value})
}
