package store

import "encoding/json"

const keyEfforts = "models.efforts"

// parseEfforts reads the last used effort per model; anything unreadable means "none remembered".
func parseEfforts(value string) map[string]string {
	efforts := map[string]string{}
	if value != "" && json.Unmarshal([]byte(value), &efforts) != nil {
		return map[string]string{}
	}
	return efforts
}

func (db *DB) SaveEfforts(efforts map[string]string) error {
	data, err := json.Marshal(efforts)
	if err != nil {
		return err
	}
	return db.setSettings(map[string]string{keyEfforts: string(data)})
}
