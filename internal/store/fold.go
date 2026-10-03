package store

import "strconv"

const (
	keyFold   = "fold"
	foldModes = 4
)

// parseFold reads the saved folding mode; anything unreadable shows everything.
func parseFold(value string) int {
	mode, err := strconv.Atoi(value)
	if err != nil || mode < 0 || mode >= foldModes {
		return 0
	}
	return mode
}

// SaveFold stores how much of the agent's work the chat shows.
func (db *DB) SaveFold(mode int) error {
	return db.setSettings(map[string]string{keyFold: strconv.Itoa(mode)})
}
