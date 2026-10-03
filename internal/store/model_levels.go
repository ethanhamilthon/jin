package store

import (
	"encoding/json"
)

// SaveModelLevels stores the reasoning levels for the active provider.
func (db *DB) SaveModelLevels(levels map[string][]string) error {
	active, err := db.ActiveProviderID()
	if err != nil {
		return err
	}
	return db.SaveModelLevelsFor(active, levels)
}

func (db *DB) SaveModelLevelsFor(providerID string, levels map[string][]string) error {
	data, err := json.Marshal(levels)
	if err != nil {
		return err
	}
	updates := map[string]string{levelsKey(providerID): string(data)}
	if providerID == "default" || providerID == "" {
		updates[keyModelsLevels] = string(data)
	}
	return db.setSettings(updates)
}

func (db *DB) LoadModelLevels() (map[string][]string, error) {
	active, err := db.ActiveProviderID()
	if err != nil {
		return nil, err
	}
	return db.LoadModelLevelsFor(active)
}

func (db *DB) LoadModelLevelsFor(providerID string) (map[string][]string, error) {
	values, err := db.settings()
	if err != nil {
		return nil, err
	}
	raw := values[levelsKey(providerID)]
	if raw == "" && (providerID == "default" || providerID == "") {
		raw = values[keyModelsLevels]
	}
	levels := map[string][]string{}
	if raw != "" && json.Unmarshal([]byte(raw), &levels) != nil {
		return map[string][]string{}, nil
	}
	return levels, nil
}
