-- 20260522120000_init.sql
-- Matches SQLite schema from tui/migrations/001_init.sql + 002_inbox.sql
-- Plus updated_at for sync conflict resolution

CREATE TABLE IF NOT EXISTS tasks (
    id             TEXT PRIMARY KEY,
    title          TEXT NOT NULL,
    description    TEXT,
    priority       TEXT NOT NULL DEFAULT 'P3',
    status         TEXT NOT NULL DEFAULT 'INBOX',
    scheduled_date TEXT,
    created_at     TEXT NOT NULL,
    completed_at   TEXT,
    updated_at     TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS notes (
    id         TEXT PRIMARY KEY,
    task_id    TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS inbox (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_tasks_status         ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_date ON tasks(scheduled_date);
CREATE INDEX IF NOT EXISTS idx_notes_task_id        ON notes(task_id);
CREATE INDEX IF NOT EXISTS idx_inbox_created_at      ON inbox(created_at);

-- Enable realtime for PowerSync change capture
ALTER TABLE tasks  REPLICA IDENTITY FULL;
ALTER TABLE notes  REPLICA IDENTITY FULL;
ALTER TABLE inbox  REPLICA IDENTITY FULL;

-- Supabase realtime publication (only create if not exists)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = 'supabase_realtime') THEN
        CREATE PUBLICATION supabase_realtime FOR ALL TABLES;
    END IF;
END;
$$;
