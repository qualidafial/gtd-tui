CREATE TABLE projects (
    id          INTEGER PRIMARY KEY,
    title       TEXT NOT NULL CHECK (title != ''),
    outcome     TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL CHECK (status IN ('active','deferred','someday','done','dropped')),
    due         DATETIME,
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE project_tasks (
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_id    INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    PRIMARY KEY (project_id, task_id)
);

CREATE INDEX idx_project_tasks_task_id ON project_tasks(task_id);
