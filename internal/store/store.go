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
	if _, err := sqlDB.Exec(schema + asyncSchema); err != nil {
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

const schema = `
CREATE TABLE IF NOT EXISTS sessions (
	id TEXT PRIMARY KEY,
	path TEXT NOT NULL,
	model TEXT NOT NULL,
	effort TEXT NOT NULL,
	title TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	input_tokens INTEGER NOT NULL DEFAULT 0,
	output_tokens INTEGER NOT NULL DEFAULT 0,
	context_tokens INTEGER NOT NULL DEFAULT 0,
	cost REAL NOT NULL DEFAULT 0,
	provider TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS sessions_path ON sessions(path, updated_at);
CREATE TABLE IF NOT EXISTS messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id TEXT NOT NULL REFERENCES sessions(id),
	data TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS messages_session_id ON messages(session_id);
CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS unread_sessions (
	session_id TEXT PRIMARY KEY REFERENCES sessions(id)
);
CREATE TABLE IF NOT EXISTS running_sessions (
	session_id TEXT PRIMARY KEY REFERENCES sessions(id),
	pid INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS todos (
	session_id TEXT NOT NULL REFERENCES sessions(id),
	position INTEGER NOT NULL,
	text TEXT NOT NULL,
	status TEXT NOT NULL,
	PRIMARY KEY (session_id, position)
);
CREATE TABLE IF NOT EXISTS todo_state (
	session_id TEXT PRIMARY KEY REFERENCES sessions(id),
	edited INTEGER NOT NULL DEFAULT 0
);
`
