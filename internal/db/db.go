// db.go
// SQLite init and schema migration

package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Init opens the SQLite database and runs schema migrations.
// dbPath is the file path, e.g. "./data/gateway.db"
func Init(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = "./data/gateway.db"
	}

	database, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Verify the connection is alive — close and wrap error on failure
	if err := database.Ping(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	// Run schema migrations — close and wrap error on failure
	if err := migrate(database); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("migrate db: %w", err)
	}

	return database, nil
}

// migrate creates all tables if they do not exist yet.
// Add new migrations as additional Exec calls below.
func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS external_users (
		id                   INTEGER PRIMARY KEY AUTOINCREMENT,
		email                TEXT    NOT NULL UNIQUE,
		name                 TEXT    NOT NULL,
		api_key_hash         TEXT    NOT NULL DEFAULT '',
		verified             INTEGER NOT NULL DEFAULT 0,
		verification_token   TEXT    NOT NULL DEFAULT '',
		verification_expires INTEGER NOT NULL DEFAULT 0,
		internal_jwt         TEXT    NOT NULL DEFAULT '',
		campaign_count       INTEGER NOT NULL DEFAULT 0,
		created_at           INTEGER NOT NULL DEFAULT (strftime('%s','now'))
	);

	CREATE INDEX IF NOT EXISTS idx_external_users_email
		ON external_users(email);

	CREATE INDEX IF NOT EXISTS idx_external_users_verification_token
		ON external_users(verification_token);
	`

	_, err := db.Exec(schema)
	return err
}
