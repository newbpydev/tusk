---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: 2026-09-06-001-feat-tusk-modern-task-system-plan.md
---

# Feature Plan 003: Task Service Engine

## Goal Capsule

Implement the application service engine in `internal/service/`. This layer coordinates domain logic with persistence: managing task creation with natural language dates, orchestrating subtask progress rollups across database boundaries, applying complex multi-criteria filters and sorting, and executing atomic bulk operations.

---

## Technical Design & Scope

### 1. Service Orchestrator (`internal/service/task_service.go`)
- Inbound interface: `ports.TaskService`.
- Methods:
  - `CreateTask(ctx, cmd CreateTaskCommand) (*core.Task, error)`
  - `UpdateTask(ctx, cmd UpdateTaskCommand) (*core.Task, error)`
  - `CompleteTask(ctx, id string) (*core.Task, error)` (triggers upward progress rollup to parents)
  - `ReopenTask(ctx, id string) (*core.Task, error)`
  - `DeleteTask(ctx, id string, recursive bool) error`
  - `ListTasks(ctx, query TaskQuery) ([]core.Task, error)`
  - `GetTaskTree(ctx, rootID string) (*core.TaskNode, error)`

### 2. Natural Language Date Parser (`internal/service/dateparse/`)
- Pure Go parser for human-friendly date strings:
  - Relative tokens: `today`, `tomorrow`, `tonight`, `mon`, `tue`, `wed`, `thu`, `fri`, `sat`, `sun`.
  - Durations: `+1d`, `+3d`, `+1w`, `+2w`, `+1m`.
  - Absolute ISO/RFC dates: `2026-09-15`, `2026-09-15T18:00:00Z`.

### 3. Subtask Rollup Orchestration
- When a task is updated or marked done/reopened, the service loads its parent hierarchy and invokes `core.CalculateProgress`.
- Updates parent progress and status atomically within a database transaction.

---

## Verification Scenarios

1. Unit tests for `dateparse` with mock reference times, ensuring timezone-accurate translations.
2. Service unit tests with mock repository verifying upward progress rollup across 3 levels of subtasks.
3. Negative tests verifying validation errors on invalid task commands without touching the repository.
4. Concurrency tests ensuring atomic updates when multiple subtasks under the same parent complete simultaneously.
