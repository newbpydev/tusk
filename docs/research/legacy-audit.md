# Legacy Technical Audit: Tusk v0.1.0 Post-Mortem

**Date**: 2026-09-06  
**Auditor**: Compound Engineering Reboot Team  
**Subject Repository**: `newbpydev/tusk` (commit `b57af46`)  
**Archival Reference**: `legacy/baseline` git tag & `docs/research/legacy-audit/snapshots/`

---

## 1. Executive Summary

Legacy Tusk was an ambitious task management application written in Go, combining a PostgreSQL storage backend, a command-line interface (CLI) using Cobra, and an interactive Terminal User Interface (TUI) powered by Charm's Bubble Tea and Lipgloss. 

While the project demonstrated commendable vision for keyboard-driven productivity, it collapsed under the weight of **structural over-engineering, tight coupling to external infrastructure, violation of Elm architecture principles in the TUI, and an extreme testing deficit**.

This document captures the forensic post-mortem analysis of the legacy codebase, establishing the technical anti-patterns that must never be repeated in the modern reboot.

---

## 2. Architecture & Layering Audit

### 2.1 Over-Engineered Hexagonal Layering
Legacy Tusk attempted an enterprise-grade hexagonal architecture (ports and adapters) with:
- `internal/core/task/model.go` (Domain model)
- `internal/ports/output/task_repository.go` (Output port)
- `internal/ports/output/user_repository.go` (User output port)
- `internal/service/task/service.go` & `implementation.go` (Use-case service)
- `internal/service/task/async_wrapper.go` (Asynchronous wrapper)
- `internal/adapters/db/task_repo.go` (PostgreSQL adapter)
- `internal/adapters/tui/bubbletea/...` (TUI adapter with over 35 files)

**Failure Mode**: For a single-user developer productivity tool, this level of indirection introduced immense friction without tangible benefit. Simple operations required tracing through 6 distinct abstraction layers.

### 2.2 The `async_wrapper.go` Concurrency & Stale State Hazard
In an attempt to prevent UI freezing during database calls, `internal/service/task/async_wrapper.go` introduced a concurrent worker pool (`worker.NewPool(10)`) and an uncoordinated `sync.Map` cache directly inside the service layer:

```go
// From snapshots/internal/service/task/async_wrapper.go (Lines 16-21)
type AsyncTaskService struct {
    taskService Service
    workerPool  *worker.Pool
    log         *zap.Logger
    cache       sync.Map // Used to cache recent operations for faster UI feedback
}
```

When listing tasks (`List` method, lines 45-67):
```go
if cachedTasks, ok := s.cache.Load("user_tasks_" + fmt.Sprintf("%d", userID)); ok {
    tasks := cachedTasks.([]task.Task)
    // Refresh in background
    s.workerPool.Submit(func() error {
        bgCtx := context.Background()
        freshTasks, err := s.taskService.List(bgCtx, userID)
        if err == nil {
            s.cache.Store("user_tasks_"+fmt.Sprintf("%d", userID), freshTasks)
        }
        return err
    })
    return tasks, nil
}
```

**Consequences**:
1. **Cache Invalidation Failure**: Mutations (creating, toggling, editing tasks) were submitted to background workers, but cache keys like `"user_tasks_" + id` were not synchronously invalidated or locked. The TUI frequently rendered stale task lists immediately after a user completed an action.
2. **Race Conditions**: Concurrent background goroutines writing to the database while reading from `sync.Map` led to non-deterministic task ordering and phantom items.
3. **Elm Architecture Violation**: Bubble Tea has its own native, safe concurrency model: `tea.Cmd`. Asynchronous operations belong in `tea.Cmd` returning a `tea.Msg` to `Update(msg)`. Placing a goroutine pool inside a domain service completely bypassed Bubble Tea's message loop.

---

## 3. Database & Authentication Coupling Audit

### 3.1 Startup Crash on `--help` or Missing PostgreSQL
In `cmd/cli/main.go`, the initialization sequence mandated a live PostgreSQL connection before parsing CLI flags or arguments:

```go
// From snapshots/cmd/cli/main.go (Lines 39-45)
ctx := context.Background()

// Initialize the database connection pool
if err := db.Connect(ctx); err != nil {
    logging.Logger.Error("Failed to connect to database", zap.Error(err))
    fmt.Println("Error: Could not connect to database. Check logs for details.")
    os.Exit(1)
}

// Execute CLI commands - services will be initialized inside
cli.Execute()
```

And in `internal/adapters/db/db.go`:
```go
// From snapshots/internal/adapters/db/db.go (Lines 17, 59-63)
var Pool *pgxpool.Pool

if err := pool.Ping(ctx); err != nil {
    Logger.Error("Failed to ping database", zap.Error(err))
    return errors.Wrap(err, "failed to ping database")
}
```

**Consequences**:
1. **Broken CLI Composability**: Running `tusk --help` or `tusk version` without a running PostgreSQL container (`docker compose up -d`) exited immediately with `exit code 1`. A developer could not even inspect help docs offline.
2. **Global Mutable State**: `var Pool *pgxpool.Pool` created a global variable accessed across packages, impeding isolated unit testing.
3. **Heavy Operational Burden**: Requiring users to install, configure, and maintain a PostgreSQL daemon just to take notes on their personal laptop created an unacceptable barrier to adoption.

### 3.2 Interactive Terminal Authentication in Batch CLI
Legacy Tusk forced a single user authentication check (`simpleTerminalAuth`) that prompted for passwords via interactive terminal input during CLI execution, making it impossible to script `tusk` commands in cron jobs, shell pipelines, or CI/CD environments.

---

## 4. TUI Architectural Anti-Patterns (Bubble Tea)

### 4.1 State Mutation Inside Pure `View()`
Bubble Tea enforces the Elm Architecture:
`Update(msg) (Model, Cmd)` is where state mutations occur.
`View() string` must be a **pure, idempotent projection** of the model to a terminal string.

In `internal/adapters/tui/bubbletea/app/view.go`, legacy Tusk repeatedly mutated the receiver model pointer `m` inside `View()`:

```go
// From snapshots/internal/adapters/tui_bubbletea/app/view.go (Lines 30-51)
func (m *Model) View() string {
    ...
    // Initialize collapsible sections if needed (MUTATION!)
    if m.collapsibleManager == nil {
        m.initCollapsibleSections()
    }
    
    // Set active keymap based on current context (MUTATION!)
    switch m.activePanel {
    case 0:
        m.activeKeyMap = keymap.TaskListKeyMap
    case 1:
        m.activeKeyMap = keymap.TaskDetailsKeyMap
    ...
    }
    
    // Update help model with current keymap (MUTATION!)
    m.helpModel.SetKeyMap(m.activeKeyMap, contextID)
    m.helpModel.AddDelegateKeyMap(keymap.GlobalKeyMap)
    m.helpModel.SetWidth(m.width)
    ...
```

**Consequences**:
- Render passes altered the internal state of the application.
- Window resize events or rapid terminal refreshes triggered nested mutation loops, leading to layout shifts, truncated borders, and severe terminal flicker.

### 4.2 Monolithic Model Struct
In `internal/adapters/tui/bubbletea/app/model.go`, a single `Model` struct spanned over 175 lines, holding:
- Raw domain data (`tasks []task.Task`)
- Viewport dimensions (`width`, `height`)
- Panel navigation indices (`activePanel`, `selectedTaskIndex`)
- Filter and search queries
- Modal dialog configurations
- Help model state
- Form input states
- Collapsible section managers
- Direct database repository references

Because state was not cleanly decomposed into hierarchical sub-models, an update to a single text input in a form required re-evaluating the entire application state tree.

---

## 5. Test Deficit Audit

### 5.1 Metrics
- **Total Go Files**: 93
- **Total Test Files**: 3
  - `internal/adapters/db/task_repo_test.go`
  - `internal/adapters/db/user_repo_test.go`
  - `internal/service/task/implementation_test.go`
- **CLI Test Coverage**: 0.0% (Zero tests for Cobra commands, flag parsing, or outputs)
- **TUI Test Coverage**: 0.0% (Zero tests for Bubble Tea `Init`, `Update`, `View`, or keymaps)
- **Integration Test Coverage**: 0.0%

### 5.2 Verification Impact
Because there were no automated regression tests:
- UI layout fixes repeatedly broke keyboard navigation handlers.
- Refactorings in the async worker pool introduced silent database deadlocks.
- The project could not be safely modified without manual verification of every UI screen.

---

## 6. Actionable Invariants for the Tusk Reboot

From this audit, five immutable design rules are established for the reboot:

1. **Zero External Dependencies by Default**: Tusk must run instantly out of the box using embedded SQLite (`~/.local/share/tusk/tusk.db`) with zero network calls, daemons, or credential setups.
2. **Pure Elm Architecture**: `View()` is strictly read-only and idempotent. All state changes, UI calculations, and dimension adjustments occur exclusively in `Update(tea.Msg)`.
3. **Graceful CLI Autonomy**: `--help`, `--version`, and configuration commands must never touch storage or attempt connections. Database initialization must be lazy and scoped to commands that require persistence.
4. **No Custom Goroutine Pools in Domain Logic**: Concurrency is handled exclusively via Bubble Tea's `tea.Cmd` in the presentation layer and standard synchronous operations in the domain/service layers.
5. **Strict Test-Driven Development (TDD)**: Every domain invariant, repository query, CLI command, and TUI component must be accompanied by comprehensive tests before implementation is merged.
