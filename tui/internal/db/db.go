package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Open opens (or creates) the SQLite database at the default config path
// and runs all pending migrations.
func Open() (*sql.DB, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	dbPath := filepath.Join(dir, "flowtask.db")
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode and foreign keys for better concurrency + referential integrity.
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA busy_timeout=5000;",
	}
	for _, p := range pragmas {
		if _, err := database.Exec(p); err != nil {
			return nil, fmt.Errorf("pragma %q: %w", p, err)
		}
	}

	if err := migrate(database); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return database, nil
}

// configDir returns ~/.config/flowtask on Linux/macOS.
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "flowtask"), nil
}

// migrate runs embedded SQL migrations in order.
// For simplicity we use a single bundled migration; extend this list as needed.
func migrate(db *sql.DB) error {
	// Create the migrations tracking table.
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version   INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		);
	`); err != nil {
		return err
	}

	migrations := []struct {
		version int
		sql     string
	}{
		{1, migration001},
		{2, migration002},
		{3, migration003},
	}

	for _, m := range migrations {
		var count int
		row := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, m.version)
		if err := row.Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue // already applied
		}

		if _, err := db.Exec(m.sql); err != nil {
			return fmt.Errorf("migration %d: %w", m.version, err)
		}
		if _, err := db.Exec(
			`INSERT INTO schema_migrations (version, applied_at) VALUES (?, datetime('now'))`,
			m.version,
		); err != nil {
			return err
		}
	}

	return nil
}

// migration001 is the initial schema (mirrors migrations/001_init.sql).
const migration001 = `
CREATE TABLE IF NOT EXISTS tasks (
    id            TEXT PRIMARY KEY,
    title         TEXT NOT NULL,
    description   TEXT,
    priority      TEXT NOT NULL DEFAULT 'P3',
    status        TEXT NOT NULL DEFAULT 'INBOX',
    scheduled_date TEXT,
    created_at    TEXT NOT NULL,
    completed_at  TEXT
);

CREATE TABLE IF NOT EXISTS notes (
    id         TEXT PRIMARY KEY,
    task_id    TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_status         ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_date ON tasks(scheduled_date);
CREATE INDEX IF NOT EXISTS idx_notes_task_id        ON notes(task_id);
`

// migration002 adds the inbox table for lightweight capture items.
const migration002 = `
CREATE TABLE IF NOT EXISTS inbox (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT,
    created_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_inbox_created_at ON inbox(created_at);
`

// migration003 adds updated_at column for sync conflict resolution.
const migration003 = `
ALTER TABLE tasks  ADD COLUMN updated_at TEXT NOT NULL DEFAULT '';
ALTER TABLE notes  ADD COLUMN updated_at TEXT NOT NULL DEFAULT '';
ALTER TABLE inbox  ADD COLUMN updated_at TEXT NOT NULL DEFAULT '';
`
