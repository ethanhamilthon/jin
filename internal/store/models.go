package store

import "encoding/json"

const (
	keyModelsCache  = "models.cache"
	keyModelsLevels = "models.levels"
)

func cacheKey(id string) string {
	if id == "" {
		return keyModelsCache
	}
	return keyModelsCache + "." + id
}

func levelsKey(id string) string {
	if id == "" {
		return keyModelsLevels
	}
	return keyModelsLevels + "." + id
}

// SaveModelsCache stores the model list for the active provider.
func (db *DB) SaveModelsCache(ids []string) error {
	active, err := db.ActiveProviderID()
	if err != nil {
		return err
	}
	return db.SaveModelsCacheFor(active, ids)
}

func (db *DB) SaveModelsCacheFor(providerID string, ids []string) error {
	data, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	updates := map[string]string{cacheKey(providerID): string(data)}
	if providerID == "default" || providerID == "" {
		updates[keyModelsCache] = string(data)
	}
	return db.setSettings(updates)
}

// LoadModelsCache returns the cached model list for active provider; nil when none.
func (db *DB) LoadModelsCache() ([]string, error) {
	active, err := db.ActiveProviderID()
	if err != nil {
		return nil, err
	}
	return db.LoadModelsCacheFor(active)
}

func (db *DB) LoadModelsCacheFor(providerID string) ([]string, error) {
	values, err := db.settings()
	if err != nil {
		return nil, err
	}
	raw := values[cacheKey(providerID)]
	if raw == "" && (providerID == "default" || providerID == "") {
		raw = values[keyModelsCache]
	}
	var ids []string
	if raw == "" || json.Unmarshal([]byte(raw), &ids) != nil {
		return nil, nil
	}
	return ids, nil
}
