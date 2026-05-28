-- Rename task status 'pending' → 'open', make assignee nullable, drop kind column.
-- Requires table recreation for CHECK constraint, column nullability, and column removal.
CREATE TABLE tasks_new (
    id                INTEGER PRIMARY KEY,
    title             TEXT NOT NULL CHECK (title != ''),
    description       TEXT NOT NULL DEFAULT '',
    status            TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'done', 'dropped')),
    assignee          TEXT,
    project_id        INTEGER REFERENCES projects(id) ON DELETE SET NULL,
    due               DATETIME,
    defer_until       DATETIME,
    order_key         TEXT,
    created_at        DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at        DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    status_changed_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO tasks_new
    (id, title, description, status, assignee, project_id, due, defer_until, order_key, created_at, updated_at, status_changed_at)
SELECT
    id, title, description,
    CASE status WHEN 'pending' THEN 'open' ELSE status END,
    NULLIF(assignee, ''),
    project_id, due, defer_until, order_key, created_at, updated_at, status_changed_at
FROM tasks;

DROP TABLE tasks;
ALTER TABLE tasks_new RENAME TO tasks;

CREATE INDEX idx_tasks_status_order_key ON tasks(status, order_key);
CREATE INDEX idx_tasks_project_id ON tasks(project_id);