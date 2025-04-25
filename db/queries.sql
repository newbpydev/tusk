-- Users ---------------------------------------------------------------
-- name: CreateUser :one
INSERT INTO users 
   (username, email, password_hash) 
VALUES 
   ($1, $2, $3)
RETURNING 
   id, username, email, created_at, updated_at, last_login, is_active;

-- name: UpdateUser :one
UPDATE users
SET 
   username = $1, email = $2, password_hash = $3, last_login = $4, is_active = $5
WHERE 
   id = $6
RETURNING 
   id, username, email, created_at, updated_at, last_login, is_active;

-- name: GetUserByUsername :one
SELECT 
   id, username, email, created_at, updated_at, last_login, is_active
FROM users 
WHERE 
   username = $1;

-- name: GetUserById :one
SELECT 
   id, username, email, created_at, updated_at, last_login, is_active
FROM users
WHERE 
   id = $1;

-- name: GetUserByEmail :one
SELECT 
   id, username, email, created_at, updated_at, last_login, is_active
FROM users
WHERE 
   email = $1;

-- name: ListUsers :many
SELECT 
   id, username, email, created_at, updated_at, last_login, is_active
FROM users
WHERE 
   is_active = true
ORDER BY
   username;

-- name: DeactivateUser :exec
UPDATE users
SET 
   is_active = false,
   updated_at = CURRENT_TIMESTAMP
WHERE 
   id = $1;

-- Tasks ---------------------------------------------------------------

-- name: CreateTask :one
INSERT INTO tasks 
   (user_id, parent_id, title, description, due_date, is_completed, status, priority, tags, display_order)
VALUES 
   ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, is_completed, status, priority, tags, display_order;

-- name: GetTaskById :one
SELECT * 
FROM tasks 
WHERE 
   id = $1;

-- name: UpdateTask :exec
UPDATE tasks
SET 
   user_id = $2, 
   parent_id = $3, 
   title = $4, 
   description = $5, 
   due_date = $6, 
   is_completed = $7, 
   status = $8, 
   priority = $9, 
   tags = $10, 
   display_order = $11
WHERE 
   id = $1
RETURNING
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, is_completed, status, priority, tags, display_order; 

-- name: DeleteTask :exec
DELETE FROM tasks 
WHERE 
   id = $1;

-- name: ListRootTasksByUserId :many
SELECT 
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, is_completed, status, priority, tags, display_order
FROM tasks
WHERE 
   user_id = $1 AND parent_id IS NULL
ORDER BY
   display_order, created_at DESC;

-- name: GetSubtasksByParentId :many
SELECT 
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, is_completed, status, priority, tags, display_order
FROM tasks
WHERE 
   parent_id = $1
ORDER BY
   display_order, created_at DESC;

-- name: ListTasksWithSubtasksRecursive :many
WITH RECURSIVE task_tree AS (
    -- Base case
    SELECT 
        t.id, t.user_id, t.parent_id, t.title, t.description, t.created_at, t.updated_at, 
        t.due_date, t.is_completed, t.status, t.priority, t.tags, t.display_order
    FROM tasks t
    WHERE t.id = $1
    
    UNION ALL
    
    -- Recursive case
    SELECT 
        t.id, t.user_id, t.parent_id, t.title, t.description, t.created_at, t.updated_at, 
        t.due_date, t.is_completed, t.status, t.priority, t.tags, t.display_order
    FROM tasks t
    INNER JOIN task_tree tt ON t.parent_id = tt.id
)
SELECT 
    task_tree.id,
    task_tree.user_id,
    task_tree.parent_id,
    task_tree.title,
    task_tree.description,
    task_tree.created_at,
    task_tree.updated_at,
    task_tree.due_date,
    task_tree.is_completed,
    task_tree.status,
    task_tree.priority,
    task_tree.tags,
    task_tree.display_order
FROM task_tree
ORDER BY task_tree.display_order, task_tree.created_at DESC;

-- name: ReorderTask :exec
UPDATE tasks
SET 
   display_order = $2
WHERE 
   id = $1;

-- Additional queries for enhanced functionality -------------------------

-- name: SearchTasksByTitle :many
SELECT 
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, 
   is_completed, status, priority, tags, display_order
FROM tasks
WHERE 
   user_id = $1 AND
   title ILIKE $2
ORDER BY
   created_at DESC;

-- name: SearchTasksByTag :many
SELECT 
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, 
   is_completed, status, priority, tags, display_order
FROM tasks
WHERE 
   user_id = $1 AND
   $2 = ANY(tags)
ORDER BY
   created_at DESC;

-- name: ListTasksByStatus :many
SELECT 
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, 
   is_completed, status, priority, tags, display_order
FROM tasks
WHERE 
   user_id = $1 AND
   status = $2
ORDER BY
   priority DESC, display_order, created_at DESC;

-- name: ListTasksByPriority :many
SELECT 
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, 
   is_completed, status, priority, tags, display_order
FROM tasks
WHERE 
   user_id = $1 AND
   priority = $2
ORDER BY
   display_order, created_at DESC;

-- name: ListTasksDueToday :many
SELECT 
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, 
   is_completed, status, priority, tags, display_order
FROM tasks
WHERE 
   user_id = $1 AND
   due_date::date = CURRENT_DATE AND
   is_completed = false
ORDER BY
   priority DESC, display_order;

-- name: ListTasksDueSoon :many
SELECT 
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, 
   is_completed, status, priority, tags, display_order
FROM tasks
WHERE 
   user_id = $1 AND
   due_date IS NOT NULL AND
   due_date::date BETWEEN CURRENT_DATE AND (CURRENT_DATE + interval '7 days')::date AND
   is_completed = false
ORDER BY
   due_date, priority DESC;

-- name: ListOverdueTasks :many
SELECT 
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, 
   is_completed, status, priority, tags, display_order
FROM tasks
WHERE 
   user_id = $1 AND
   due_date < CURRENT_DATE AND
   is_completed = false
ORDER BY
   due_date, priority DESC;

-- name: GetTaskCountsByStatus :one
SELECT
   COUNT(*) FILTER (WHERE status = 'todo') AS todo_count,
   COUNT(*) FILTER (WHERE status = 'in-progress') AS in_progress_count,
   COUNT(*) FILTER (WHERE status = 'done') AS done_count,
   COUNT(*) AS total_count
FROM tasks
WHERE
   user_id = $1;

-- name: GetTaskCountsByPriority :one
SELECT
   COUNT(*) FILTER (WHERE priority = 'low') AS low_priority_count,
   COUNT(*) FILTER (WHERE priority = 'medium') AS medium_priority_count,
   COUNT(*) FILTER (WHERE priority = 'high') AS high_priority_count
FROM tasks
WHERE
   user_id = $1 AND
   is_completed = false;

-- name: GetRecentlyCompletedTasks :many
SELECT 
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, 
   is_completed, status, priority, tags, display_order
FROM tasks
WHERE 
   user_id = $1 AND
   is_completed = true
ORDER BY
   updated_at DESC
LIMIT $2;

-- name: BulkUpdateTaskStatus :exec
UPDATE tasks
SET 
   status = $2,
   is_completed = $3,
   updated_at = CURRENT_TIMESTAMP
WHERE 
   id = ANY($1::int[]);

-- name: GetAllTagsForUser :many
SELECT DISTINCT unnest(tags) as tag
FROM tasks
WHERE user_id = $1
ORDER BY tag;



