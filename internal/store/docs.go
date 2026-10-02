package store

const keyJinDocs = "docs.enabled"

// SaveJinDocs stores whether new sessions get the pointer to the jin docs.
func (db *DB) SaveJinDocs(enabled bool) error {
	value := "0"
	if enabled {
		value = "1"
	}
	return db.setSettings(map[string]string{keyJinDocs: value})
}
