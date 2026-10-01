package store

import "encoding/json"

const keyScope = "models.scope"

// parseScope reads the saved model scope; anything unreadable means "no scope".
func parseScope(value string) []string {
	var scope []string
	if value == "" || json.Unmarshal([]byte(value), &scope) != nil {
		return nil
	}
	return scope
}

// SaveScope stores the models offered in the model picker. An empty scope
// means every model is offered.
func (db *DB) SaveScope(scope []string) error {
	value := ""
	if len(scope) > 0 {
		data, err := json.Marshal(scope)
		if err != nil {
			return err
		}
		value = string(data)
	}
	return db.setSettings(map[string]string{keyScope: value})
}
