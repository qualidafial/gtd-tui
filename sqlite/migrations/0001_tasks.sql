CREATE TABLE tasks (
    id           INTEGER PRIMARY KEY,
    title        TEXT NOT NULL CHECK (title != ''),
    description  TEXT NOT NULL DEFAULT '',
    kind         TEXT NOT NULL DEFAULT 'next_action' CHECK (kind IN ('next_action', 'delegated')),
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'done', 'dropped')),
    assignee     TEXT NOT NULL DEFAULT '',
    due          DATETIME,
    defer_until  DATETIME,
    order_key    TEXT,
    created_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_tasks_status_order_key ON tasks(status, order_key);
