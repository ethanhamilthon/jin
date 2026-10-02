package store

import "encoding/json"

const (
	keyModelsCache  = "models.cache"
	keyModelsLevels = "models.levels"
)

// SaveModelsCache stores the model list fetched by `jin refresh-models`.
func (db *DB) SaveModelsCache(ids []string) error {
	data, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	return db.setSettings(map[string]string{keyModelsCache: string(data)})
}

// LoadModelsCache returns the cached model list; nil when there is none.
func (db *DB) LoadModelsCache() ([]string, error) {
	values, err := db.settings()
	if err != nil {
		return nil, err
	}
	var ids []string
	if values[keyModelsCache] == "" || json.Unmarshal([]byte(values[keyModelsCache]), &ids) != nil {
		return nil, nil
	}
	return ids, nil
}

// SaveModelLevels stores the reasoning levels per model. A model without an
// entry has unknown levels.
func (db *DB) SaveModelLevels(levels map[string][]string) error {
	data, err := json.Marshal(levels)
	if err != nil {
		return err
	}
	return db.setSettings(map[string]string{keyModelsLevels: string(data)})
}

func (db *DB) LoadModelLevels() (map[string][]string, error) {
	values, err := db.settings()
	if err != nil {
		return nil, err
	}
	levels := map[string][]string{}
	if values[keyModelsLevels] != "" && json.Unmarshal([]byte(values[keyModelsLevels]), &levels) != nil {
		return map[string][]string{}, nil
	}
	return levels, nil
}
