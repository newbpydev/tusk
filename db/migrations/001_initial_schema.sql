CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY CHECK (version > 0),
    filename TEXT NOT NULL UNIQUE,
    checksum TEXT NOT NULL CHECK (length(checksum) = 64 AND checksum NOT GLOB '*[^0-9a-f]*'),
    applied_at TEXT NOT NULL
) STRICT;

CREATE TABLE tasks (
    id TEXT PRIMARY KEY NOT NULL CHECK (length(id) > 0),
    title TEXT NOT NULL CHECK (length(title) BETWEEN 1 AND 255),
    description TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('todo', 'in-progress', 'blocked', 'done')),
    priority INTEGER NOT NULL CHECK (priority BETWEEN 1 AND 4),
    progress INTEGER NOT NULL CHECK (progress BETWEEN 0 AND 100),
    parent_id TEXT REFERENCES tasks(id) ON DELETE CASCADE CHECK (parent_id <> id),
    tags TEXT NOT NULL CHECK (json_valid(tags) AND json_type(tags) = 'array'),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    due_date TEXT,
    completed_at TEXT,
    CHECK ((status = 'done' AND progress = 100 AND completed_at IS NOT NULL)
        OR (status <> 'done' AND completed_at IS NULL))
) STRICT;

CREATE INDEX tasks_status_idx ON tasks(status);
CREATE INDEX tasks_priority_idx ON tasks(priority);
CREATE INDEX tasks_parent_idx ON tasks(parent_id);
CREATE INDEX tasks_due_idx ON tasks(due_date);

CREATE TABLE task_events (
    sequence INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('create', 'metadata', 'status', 'move', 'progress', 'rollup')),
    changed_fields TEXT NOT NULL CHECK (json_valid(changed_fields) AND json_type(changed_fields) = 'array' AND json_array_length(changed_fields) > 0),
    occurred_at TEXT NOT NULL
) STRICT;

CREATE INDEX task_events_task_sequence_idx ON task_events(task_id, sequence);
