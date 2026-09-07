---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: 2026-09-06-001-feat-tusk-modern-task-system-plan.md
---

# Feature Plan 002: SQLite Storage & Repository

## Goal Capsule

Implement embedded, zero-configuration local persistence using pure Go SQLite (`modernc.org/sqlite` or `ncruces/go-sqlite3`). Generate type-safe SQL queries with `sqlc` v2, package embedded migrations with `io/fs`, and implement the `ports.TaskRepository` interface with full CRUD and tree querying capabilities.

---

## Technical Design & Scope

### 1. Database Schema & Migrations (`db/migrations/`)
- `001_initial_schema.sql`:
  - `tasks` table: `id` (TEXT PRIMARY KEY), `title` (TEXT NOT NULL), `description` (TEXT), `status` (TEXT NOT NULL), `priority` (INTEGER NOT NULL), `parent_id` (TEXT REFERENCES tasks(id) ON DELETE CASCADE), `progress` (INTEGER NOT NULL DEFAULT 0), `tags` (TEXT NOT NULL DEFAULT '[]'), `due_date` (DATETIME), `created_at` (DATETIME NOT NULL), `updated_at` (DATETIME NOT NULL), `completed_at` (DATETIME).
  - Indexes: `idx_tasks_status`, `idx_tasks_priority`, `idx_tasks_parent_id`, `idx_tasks_due_date`.
- Migration Engine: Embedded Go migrations executed within an atomic transaction at application initialization.

### 2. Query Generation via `sqlc` v2 (`db/queries.sql`)
- Queries:
  - `CreateTask`, `GetTaskByID`, `UpdateTask`, `DeleteTask`.
  - `ListTasks` (with dynamic filter conditions for status, priority, due date).
  - `ListSubtasks` (retrieving immediate children by `parent_id`).
  - `GetTaskSubtree` (recursive Common Table Expression - CTE to retrieve entire child hierarchies).

### 3. Repository Implementation (`internal/storage/`)
- `SQLiteTaskRepository`: Concrete implementation of `ports.TaskRepository`.
- Connection Management: Configures WAL mode, busy timeouts, and serialized single-writer / pooled-reader handles.
- In-memory testing support using URI `file::memory:?cache=shared`.

---

## Verification Scenarios

1. Migration tests: verify schema creates successfully and rollback drops all objects cleanly.
2. In-memory integration tests: execute full CRUD lifecycle on tasks.
3. Hierarchical CTE tests: verify `GetTaskSubtree` correctly retrieves multi-level subtask trees.
4. Concurrent read/write stress test: verify WAL mode handles concurrent goroutine reads without locking errors.
