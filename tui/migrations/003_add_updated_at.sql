-- migrations/003_add_updated_at.sql

ALTER TABLE tasks  ADD COLUMN updated_at TEXT NOT NULL DEFAULT '';
ALTER TABLE notes  ADD COLUMN updated_at TEXT NOT NULL DEFAULT '';
ALTER TABLE inbox  ADD COLUMN updated_at TEXT NOT NULL DEFAULT '';
