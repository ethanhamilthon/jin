package store

import "encoding/json"

const keyScope = "models.scope"

func scopeKey(id string) string {
	if id == "" {
		return keyScope
	}
	return keyScope + "." + id
}

// parseScope reads the saved model scope; anything unreadable means "no scope".
func parseScope(value string) []string {
	var scope []string
	if value == "" || json.Unmarshal([]byte(value), &scope) != nil {
		return nil
	}
	return scope
}

// SaveScope stores the models offered for the active provider.
func (db *DB) SaveScope(scope []string) error {
	active, err := db.ActiveProviderID()
	if err != nil {
		return err
	}
	return db.SaveScopeFor(active, scope)
}

func (db *DB) SaveScopeFor(providerID string, scope []string) error {
	value := ""
	if len(scope) > 0 {
		data, err := json.Marshal(scope)
		if err != nil {
			return err
		}
		value = string(data)
	}
	updates := map[string]string{scopeKey(providerID): value}
	if providerID == "default" || providerID == "" {
		updates[keyScope] = value
	}
	return db.setSettings(updates)
}

func (db *DB) LoadScope() ([]string, error) {
	active, err := db.ActiveProviderID()
	if err != nil {
		return nil, err
	}
	return db.LoadScopeFor(active)
}

func (db *DB) LoadScopeFor(providerID string) ([]string, error) {
	values, err := db.settings()
	if err != nil {
		return nil, err
	}
	raw := values[scopeKey(providerID)]
	if raw == "" && (providerID == "default" || providerID == "") {
		raw = values[keyScope]
	}
	return parseScope(raw), nil
}
