-- Add status_changed_at by rebuilding the table: a fresh CREATE TABLE permits
-- the same non-constant default as created_at/updated_at (ADD COLUMN does not on
-- a non-empty table). Existing rows backfill from updated_at via the SELECT.
CREATE TABLE tasks_new (
    id                INTEGER PRIMARY KEY,
    title             TEXT NOT NULL CHECK (title != ''),
    description       TEXT NOT NULL DEFAULT '',
    kind              TEXT NOT NULL DEFAULT 'next_action' CHECK (kind IN ('next_action', 'delegated')),
    status            TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'done', 'dropped')),
    assignee          TEXT NOT NULL DEFAULT '',
    due               DATETIME,
    defer_until       DATETIME,
    order_key         TEXT,
    created_at        DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at        DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    status_changed_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO tasks_new
    (id, title, description, kind, status, assignee, due, defer_until, order_key, created_at, updated_at, status_changed_at)
SELECT
    id, title, description, kind, status, assignee, due, defer_until, order_key, created_at, updated_at, updated_at
FROM tasks;

DROP TABLE tasks;
ALTER TABLE tasks_new RENAME TO tasks;

CREATE INDEX idx_tasks_status_order_key ON tasks(status, order_key);
