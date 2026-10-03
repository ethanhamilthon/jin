package store

import (
	_ "modernc.org/sqlite"
)

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
