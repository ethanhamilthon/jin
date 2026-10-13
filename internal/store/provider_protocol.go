package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

const keyProtocols = "provider.protocols"

func normalizeProviderURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimRight(strings.TrimSpace(raw), "/")
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawPath = ""
	return u.String()
}

func protocolKey(id, baseURL, model string) string {
	parts := []string{id, normalizeProviderURL(baseURL), model}
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(strconv.Itoa(len(part)))
		b.WriteByte(':')
		b.WriteString(part)
	}
	return b.String()
}

func (db *DB) LoadProtocol(id, baseURL, model string) (string, error) {
	list, _, err := db.LoadProviders()
	if err != nil {
		return "", err
	}
	entry, ok := findProvider(list, id)
	if !ok || entry.Removed() || entry.Kind != "auto" || normalizeProviderURL(entry.BaseURL) != normalizeProviderURL(baseURL) {
		return "", nil
	}
	values, err := db.settings()
	if err != nil {
		return "", err
	}
	var choices map[string]string
	if err := json.Unmarshal([]byte(values[keyProtocols]), &choices); err != nil && values[keyProtocols] != "" {
		return "", err
	}
	return choices[protocolKey(id, baseURL, model)], nil
}

func (db *DB) SaveProtocol(id, baseURL, model, protocol string) error {
	if protocol != "openai" && protocol != "responses" {
		return errors.New("invalid protocol")
	}
	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var raw string
	if err = tx.QueryRow(`SELECT value FROM settings WHERE key=?`, keyProviders).Scan(&raw); err != nil {
		return err
	}
	entry, ok := findProvider(parseProviders(raw), id)
	if !ok || entry.Removed() || entry.Kind != "auto" || normalizeProviderURL(entry.BaseURL) != normalizeProviderURL(baseURL) {
		return errors.New("stale provider protocol save")
	}
	var choices map[string]string
	var stored string
	err = tx.QueryRow(`SELECT value FROM settings WHERE key=?`, keyProtocols).Scan(&stored)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if stored != "" {
		if err = json.Unmarshal([]byte(stored), &choices); err != nil {
			return err
		}
	}
	if choices == nil {
		choices = map[string]string{}
	}
	choices[protocolKey(id, baseURL, model)] = protocol
	data, err := json.Marshal(choices)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, keyProtocols, string(data)); err != nil {
		return err
	}
	return tx.Commit()
}
