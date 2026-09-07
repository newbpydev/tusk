# Constraints, Invariants & Architectural Insights

**Date**: 2026-09-06  
**Status**: Core Architectural Reference  
**Audience**: Tusk Core Contributors & AI Agents

---

## 1. Domain Model Invariants

### 1.1 Task States & Valid Transitions
A task must exist in one of four distinct states:
1. `todo` (Default state upon creation)
2. `in-progress` (Task is actively being worked on)
3. `blocked` (Task cannot proceed due to external dependencies)
4. `done` (Task has been successfully completed)

**Transition Rules**:
- A task can transition between any valid state, provided business constraints are met.
- When transitioning to `done`, `completed_at` must be stamped with the current UTC timestamp.
- If a task in `done` state is reopened (moved to `todo` or `in-progress`), `completed_at` must be reset to `NULL`.

### 1.2 Recursive Subtask Hierarchy & Invariants
- A task may have zero or more direct child tasks (subtasks).
- **Depth Ceiling**: Subtask hierarchies are supported recursively up to an arbitrary n-level depth, with a recommended UI display limit of 5 levels to avoid terminal overflow.
- **Acyclic Invariant**: A task can never be its own parent or descendant. Tree cycles are strictly rejected with `ErrCyclicDependency`.
- **Automatic Progress Rollup**:
  - A task with no subtasks computes progress as:
    $$\text{Progress} = \begin{cases} 100\% & \text{if status} = \text{done} \\ 0\% & \text{otherwise} \end{cases}$$
  - A task with subtasks computes progress as the weighted or arithmetic mean of its direct children:
    $$\text{Progress} = \left\lfloor \frac{\sum_{i=1}^{N} \text{subtask}_i.\text{Progress}}{N} \right\rfloor$$
  - When all direct children of a parent task reach `done`, the parent may optionally auto-complete based on configuration (`auto_complete_parent: true`).
  - If any child of a `done` parent is reopened, the parent state must revert from `done` to `in-progress` or `todo`.

### 1.3 Priority Scale
Priority is strictly typed:
- `urgent` (Weight: 4, Color: Red / ANSI 196)
- `high` (Weight: 3, Color: Orange / ANSI 208)
- `medium` (Weight: 2, Color: Yellow / ANSI 220)
- `low` (Weight: 1, Color: Gray / ANSI 245)

### 1.4 Natural Language Due Dates
Due dates support standard RFC3339 timestamps as well as developer-friendly natural language inputs:
- `today`, `tomorrow`, `tonight`
- Day of week: `mon`, `tue`, `wed`, `thu`, `fri`, `sat`, `sun`
- Offsets: `+1d`, `+3d`, `+1w`, `+2w`, `+1m`
- Specific dates: `2026-09-15`, `Sep 15`

---

## 2. Developer Experience (DX) Requirements

### 2.1 Unix Pipeline Composability
- Every CLI subcommand that outputs data must support a clean machine-readable format via `--json`.
- Standard output (`stdout`) is reserved strictly for requested data; informational messages, spinners, and log warnings are routed exclusively to standard error (`stderr`).
- Exit codes must strictly communicate status:
  - `0`: Success
  - `1`: General error (e.g., entity not found, validation failure)
  - `2`: Invalid CLI syntax or flag parsing error

### 2.2 Terminal User Interface (TUI) Standards
- **Zero Layout Shifts**: Panels, borders, and footer help bars must maintain fixed or proportionally calculated dimensions. Adding or removing tasks must never cause the viewport borders to jump.
- **Vim Navigation**: Full support for standard modal navigation (`h`/`j`/`k`/`l`, `/` for search, `a` for add, `x` for toggle done, `d` for delete, `e` for edit).
- **Responsive Terminal Resizing**: Terminal resize events (`tea.WindowSizeMsg`) must dynamically recalculate panel widths and heights while maintaining minimum viable dimensions (e.g., 80x24 characters).

---

## 3. Storage & Concurrency Invariants

### 3.1 Single-Writer, Multi-Reader Concurrency
- SQLite under WAL mode supports concurrent readers, but serializes writers.
- All database write operations must use connection-level timeouts and transactions (`BEGIN IMMEDIATE`) to prevent busy lock errors (`SQLITE_BUSY`).
- The database connection pool must configure:
  - Max Open Connections: `1` for writer or pooled with connection timeouts for readers.
  - Busy Timeout: 5000ms (`PRAGMA busy_timeout = 5000;`).

### 3.2 Migration Safety
- Database schema changes are executed strictly through sequential migration files (`001_initial.sql`, `002_add_index.sql`).
- Migrations are idempotent and run automatically upon first database access.
- Embedded migration files (`//go:embed migrations/*.sql`) ensure the compiled binary remains completely self-contained.
