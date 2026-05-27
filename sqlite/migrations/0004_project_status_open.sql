-- Rename project status 'active' → 'open' and fix the CHECK constraint.
-- For DBs created after 0003 was corrected this is a no-op.
UPDATE projects SET status = 'open' WHERE status = 'active';

CREATE TABLE projects_new (
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

INSERT INTO projects_new SELECT * FROM projects;

DROP TABLE projects;

ALTER TABLE projects_new RENAME TO projects;