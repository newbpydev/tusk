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

-- Tasks ---------------------------------------------------------------

-- name: CreateTask :one
INSERT INTO tasks 
   (user_id, parent_id, title, description, due_date, is_completed, status, priority, tags, display_order)
VALUES 
   ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, is_completed, status, priority, tags, display_order;

-- name: GetTaskById :one
SELECT 
   id, user_id, parent_id, title, description, created_at, updated_at, due_date, is_completed, status, priority, tags, display_order
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



