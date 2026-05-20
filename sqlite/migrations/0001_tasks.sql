CREATE TABLE tasks (
    id           INTEGER PRIMARY KEY,
    title        TEXT NOT NULL CHECK (title != ''),
    description  TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL CHECK (status IN ('inbox','active','waiting','deferred','done','dropped')),
    due          DATETIME,
    defer_until  DATETIME,
    created_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

