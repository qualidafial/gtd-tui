CREATE TABLE projects (
    id                INTEGER PRIMARY KEY,
    title             TEXT NOT NULL CHECK (title != ''),
    outcome           TEXT NOT NULL DEFAULT '',
    description       TEXT NOT NULL DEFAULT '',
    due               DATETIME,
    status            TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','someday','done','dropped')),
    order_key         TEXT,
    created_at        DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at        DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    status_changed_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- A nullable FK column defaulting to NULL is addable in place; ON DELETE SET
-- NULL detaches tasks if a project is ever hard-deleted (projects are normally
-- only transitioned, not deleted).
ALTER TABLE tasks ADD COLUMN project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL;

CREATE INDEX idx_tasks_project_id ON tasks(project_id);
