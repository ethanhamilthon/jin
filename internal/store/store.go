// Package store persists sessions and settings in ~/.jin/jin.db. A session
// row is written lazily, on the first prompt, so idle chats leave no trace.
package store

import (
	"database/sql"
	"net/url"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"jin/internal/paths"
)

type DB struct {
	sql *sql.DB
}

func Open() (*DB, error) {
	file, err := paths.Global("jin.db")
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		return nil, err
	}
	if err := secureDBFiles(file); err != nil {
		return nil, err
	}
	sqlDB, err := sql.Open("sqlite", dsn(file))
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	if _, err := sqlDB.Exec(schema + asyncSchema + changesSchema); err != nil {
		sqlDB.Close()
		return nil, err
	}
	if err := migrate(sqlDB); err != nil {
		sqlDB.Close()
		return nil, err
	}
	if err := secureDBFiles(file); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return &DB{sql: sqlDB}, nil
}

// dsn lets several jin processes share the database: WAL keeps readers and a
// writer from blocking each other, busy_timeout makes a second writer wait its
// turn instead of failing, and immediate transactions never fail halfway when
// they upgrade from reading to writing.
func dsn(file string) string {
	query := url.Values{}
	query.Add("_pragma", "journal_mode(WAL)")
	query.Add("_pragma", "busy_timeout(10000)")
	query.Set("_txlock", "immediate")
	return (&url.URL{Scheme: "file", Path: file, RawQuery: query.Encode()}).String()
}

func (db *DB) Close() error {
	return db.sql.Close()
}
