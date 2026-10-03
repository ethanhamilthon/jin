package store

import (
	"encoding/json"
	"path/filepath"
)

const keyHooksTrust = "hooks.trust"

// Trust is the answer for the project hooks of one directory.
type Trust int

const (
	TrustUnknown Trust = iota
	Trusted
	Distrusted
)

func (db *DB) trustMap() (map[string]bool, error) {
	values, err := db.settings()
	if err != nil {
		return nil, err
	}
	trust := map[string]bool{}
	if raw := values[keyHooksTrust]; raw != "" {
		_ = json.Unmarshal([]byte(raw), &trust)
	}
	return trust, nil
}

func trustKey(dir string) string {
	if abs, err := filepath.Abs(dir); err == nil {
		return abs
	}
	return dir
}

// HooksTrust says whether the project hooks of dir may run.
func (db *DB) HooksTrust(dir string) (Trust, error) {
	trust, err := db.trustMap()
	if err != nil {
		return TrustUnknown, err
	}
	answer, ok := trust[trustKey(dir)]
	switch {
	case !ok:
		return TrustUnknown, nil
	case answer:
		return Trusted, nil
	}
	return Distrusted, nil
}

// SaveHooksTrust records the answer for the project hooks of dir.
func (db *DB) SaveHooksTrust(dir string, trusted bool) error {
	trust, err := db.trustMap()
	if err != nil {
		return err
	}
	trust[trustKey(dir)] = trusted
	data, err := json.Marshal(trust)
	if err != nil {
		return err
	}
	return db.setSettings(map[string]string{keyHooksTrust: string(data)})
}
