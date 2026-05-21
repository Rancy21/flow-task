-- migrations/002_inbox.sql

CREATE TABLE IF NOT EXISTS inbox (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT,
    created_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_inbox_created_at ON inbox(created_at);
