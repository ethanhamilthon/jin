package store

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

// Device is a browser that jin web let in. Only a hash of its token is stored.
type Device struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Created  int64  `json:"created"`
	LastSeen int64  `json:"lastSeen"`
}

func randomHex(n int) string {
	raw := make([]byte, n)
	_, _ = rand.Read(raw)
	return hex.EncodeToString(raw)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// AddDevice registers a device and returns its secret token, which is not
// stored and cannot be read again.
func (db *DB) AddDevice(name string) (Device, string, error) {
	now := time.Now().Unix()
	device := Device{ID: randomHex(8), Name: name, Created: now, LastSeen: now}
	token := randomHex(24)
	err := retryBusy(func() error {
		_, err := db.sql.Exec(`INSERT INTO devices(id, token_hash, name, created_at, last_seen_at) VALUES (?, ?, ?, ?, ?)`,
			device.ID, hashToken(token), name, now, now)
		return err
	})
	return device, token, err
}

// DeviceByToken finds the device a token belongs to.
func (db *DB) DeviceByToken(token string) (Device, bool) {
	var d Device
	err := db.sql.QueryRow(`SELECT id, name, created_at, last_seen_at FROM devices WHERE token_hash = ?`,
		hashToken(token)).Scan(&d.ID, &d.Name, &d.Created, &d.LastSeen)
	return d, err == nil
}

func (db *DB) Devices() ([]Device, error) {
	rows, err := db.sql.Query(`SELECT id, name, created_at, last_seen_at FROM devices ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	devices := []Device{}
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.ID, &d.Name, &d.Created, &d.LastSeen); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (db *DB) TouchDevice(id string) {
	_ = retryBusy(func() error {
		_, err := db.sql.Exec(`UPDATE devices SET last_seen_at = ? WHERE id = ?`, time.Now().Unix(), id)
		return err
	})
}

func (db *DB) RenameDevice(id, name string) error {
	if name == "" {
		return errors.New("a device needs a name")
	}
	return db.changeDevice(`UPDATE devices SET name = ? WHERE id = ?`, name, id)
}

func (db *DB) RemoveDevice(id string) error {
	return db.changeDevice(`DELETE FROM devices WHERE id = ?`, id)
}

func (db *DB) changeDevice(query string, args ...any) error {
	return retryBusy(func() error {
		res, err := db.sql.Exec(query, args...)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return sql.ErrNoRows
		}
		return nil
	})
}
