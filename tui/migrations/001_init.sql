-- migrations/001_init.sql

CREATE TABLE IF NOT EXISTS tasks (
    id            TEXT PRIMARY KEY,
    title         TEXT NOT NULL,
    description   TEXT,
    priority      TEXT NOT NULL DEFAULT 'P3', -- P1 | P2 | P3
    status        TEXT NOT NULL DEFAULT 'INBOX', -- INBOX | SCHEDULED | DONE
    scheduled_date TEXT,                      -- ISO date 'YYYY-MM-DD', NULL = inbox
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
