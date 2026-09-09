-- name: GetTask :one
SELECT * FROM tasks WHERE id = ?;

-- name: CreateTask :exec
INSERT INTO tasks (id,title,description,status,priority,progress,parent_id,tags,created_at,updated_at,due_date,completed_at)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?);

-- name: UpdateTask :execrows
UPDATE tasks SET title=?,description=?,status=?,priority=?,progress=?,parent_id=?,tags=?,updated_at=?,due_date=?,completed_at=? WHERE id=?;

-- name: DeleteTask :execrows
DELETE FROM tasks WHERE id=?;

-- name: ListCandidates :many
SELECT * FROM tasks
WHERE (json_array_length(CAST(sqlc.arg(statuses) AS TEXT))=0 OR status IN (SELECT value FROM json_each(CAST(sqlc.arg(statuses) AS TEXT))))
AND (json_array_length(CAST(sqlc.arg(priorities) AS TEXT))=0 OR priority IN (SELECT value FROM json_each(CAST(sqlc.arg(priorities) AS TEXT))))
AND (CAST(sqlc.arg(root_only) AS INTEGER)=0 OR parent_id IS NULL)
AND (CAST(sqlc.narg(parent_id) AS TEXT) IS NULL OR parent_id=CAST(sqlc.narg(parent_id) AS TEXT))
AND (CAST(sqlc.narg(due_before) AS TEXT) IS NULL OR due_date < CAST(sqlc.narg(due_before) AS TEXT))
AND (CAST(sqlc.narg(due_after) AS TEXT) IS NULL OR due_date > CAST(sqlc.narg(due_after) AS TEXT));

-- name: ListChildren :many
SELECT * FROM tasks WHERE parent_id=?;

-- name: GetSubtree :many
WITH RECURSIVE subtree(id) AS (
    SELECT tasks.id FROM tasks WHERE tasks.id=?
    UNION
    SELECT tasks.id FROM tasks JOIN subtree ON tasks.parent_id=subtree.id
)
SELECT tasks.* FROM tasks JOIN subtree ON tasks.id=subtree.id;

-- name: GetAncestors :many
WITH RECURSIVE ancestors(id) AS (
    SELECT tasks.parent_id FROM tasks WHERE tasks.id=? AND tasks.parent_id IS NOT NULL
    UNION
    SELECT tasks.parent_id FROM tasks JOIN ancestors ON tasks.id=ancestors.id WHERE tasks.parent_id IS NOT NULL
)
SELECT tasks.* FROM tasks JOIN ancestors ON tasks.id=ancestors.id;

-- name: AppendEvent :one
INSERT INTO task_events(task_id,kind,changed_fields,occurred_at) VALUES(?,?,?,?) RETURNING sequence;

-- name: ListEvents :many
SELECT * FROM task_events WHERE task_id=? ORDER BY sequence;
