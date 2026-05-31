CREATE TABLE items (
    id                        INTEGER PRIMARY KEY,
    title                     TEXT NOT NULL CHECK (title != ''),
    description               TEXT NOT NULL DEFAULT '',
    created_at                DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at                DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    clarified_into_task_id    INTEGER REFERENCES tasks(id)    ON DELETE SET NULL,
    clarified_into_project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL,
    discarded                 INTEGER NOT NULL DEFAULT 0 CHECK (discarded IN (0, 1)),
    -- At most one terminal state may be set: clarified into task, clarified
    -- into project, or discarded. implement-references extends this constraint
    -- to include clarified_into_reference_id.
    CHECK (
        (CASE WHEN clarified_into_task_id    IS NOT NULL THEN 1 ELSE 0 END) +
        (CASE WHEN clarified_into_project_id IS NOT NULL THEN 1 ELSE 0 END) +
        discarded <= 1
    )
);

CREATE INDEX idx_items_inbox ON items(created_at)
    WHERE clarified_into_task_id IS NULL
      AND clarified_into_project_id IS NULL
      AND discarded = 0;