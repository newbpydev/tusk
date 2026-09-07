# Modern Stack & Technical Strategy (2026)

**Date**: 2026-09-06  
**Status**: Approved Architecture Baseline  
**Audience**: Tusk Core Contributors & AI Agents

---

## 1. Technical Vision

Tusk 2026 is engineered as a high-performance, single-binary CLI and TUI task management system. The architecture prioritizes:
- **Zero-Friction Startup**: Initial run requires no docker containers, no external daemons, and no network access.
- **Instant Response Times**: CLI subcommands execute in under 10 milliseconds; TUI responds at 60 FPS without layout jitter.
- **Predictable State Mechanics**: Strict unidirectional data flow modeled after the Elm Architecture.
- **Verifiable Reliability**: 100% test-driven coverage for domain logic, repository operations, CLI commands, and TUI view states.

---

## 2. Core Technology Choices

### 2.1 Language & Toolchain: Go 1.24+
- **Toolchain**: Go 1.24+ standard library.
- **Standard Library Primitives**:
  - `context` for deadline propagation and graceful cancellation.
  - `errors.Is` / `errors.As` with strongly typed domain error sentinels.
  - `time` with timezone-aware natural date parsing.
  - `sync` for safe concurrent operations where explicitly necessary.
- **Directory Layout Standard**:
  ```
  tusk/
  ├── cmd/
  │   └── tusk/             # Main application entrypoint
  ├── internal/
  │   ├── core/             # Pure domain entities, business invariants, errors
  │   ├── ports/            # Inbound (service) and outbound (repo) interfaces
  │   ├── storage/          # SQLite migrations, sqlc generated queries, repo adapter
  │   ├── service/          # Orchestration, filtering, rollup, business operations
  │   ├── cli/              # Cobra commands, flag binding, formatters (JSON/table)
  │   └── tui/              # Bubble Tea models, components, themes, keymaps
  ├── db/                   # Canonical schema.sql and queries.sql
  ├── scripts/              # Validation, formatting, and setup scripts
  └── docs/                 # Research, plans, and architectural guidelines
  ```

### 2.2 Storage Strategy: Embedded SQLite by Default
- **Engine**: Pure Go SQLite driver (`modernc.org/sqlite` or `ncruces/go-sqlite3`).
  - **Zero CGO**: Ensures seamless cross-compilation across Linux (x86_64, aarch64), macOS (Intel, Apple Silicon), and Windows.
  - **File Location**: Conforms to XDG Base Directory Specification (`$XDG_DATA_HOME/tusk/tusk.db` or `~/.local/share/tusk/tusk.db`).
  - **Concurrency & Durability**: WAL mode (`PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL; PRAGMA foreign_keys=ON;`), enabling high concurrent read throughput and atomic writes without database locks.
- **Pluggable Interface**:
  The domain interacts solely with `ports.TaskRepository`:
  ```go
  type TaskRepository interface {
      Create(ctx context.Context, task *core.Task) error
      GetByID(ctx context.Context, id string) (*core.Task, error)
      Update(ctx context.Context, task *core.Task) error
      Delete(ctx context.Context, id string) error
      List(ctx context.Context, filter core.TaskFilter) ([]core.Task, error)
  }
  ```
  This guarantees that an optional PostgreSQL adapter (`pgx/v5`) can be introduced for enterprise sync without modifying any domain or service code.

### 2.3 Type-Safe Persistence: `sqlc` v2
- **Philosophy**: Write pure SQL; generate idiomatic, compile-time verified Go structs and methods.
- **Elimination of Reflection**: Unlike heavy ORMs (GORM, Ent), `sqlc` generates lightweight Go code with zero runtime overhead.
- **Engine Configuration**:
  ```yaml
  version: "2"
  sql:
    - schema: "db/migrations"
      queries: "db/queries.sql"
      gen:
        go:
          package: "storage"
          out: "internal/storage"
          sql_package: "database/sql"
          emit_json_tags: true
          emit_prepared_queries: false
          emit_interface: true
  ```

### 2.4 Terminal Presentation: Charm Suite (Bubble Tea v1.3+, Lipgloss v1.1+, Bubbles v0.21+)
- **Bubble Tea v1.3+**: Unidirectional state machine:
  - `Init() tea.Cmd`: Returns initial async commands (e.g., load tasks from repository).
  - `Update(tea.Msg) (tea.Model, tea.Cmd)`: Handles all messages (key presses, window resizes, data loaded) and returns updated state.
  - `View() string`: Pure string projection. Mutating any field of the model inside `View()` is strictly prohibited.
- **Lipgloss v1.1+**: Declarative terminal layout and adaptive theming (ANSI 16-color fallback, 256-color, and True Color detection).
- **Bubbles v0.21+**: Reusable UI components (textinput for task entry, list for task browsing, viewport for task notes, help for keymap visualization).

---

## 3. Testing Harnesses & Quality Verification

### 3.1 Domain Invariant Testing
- 100% unit test coverage for `internal/core/` and `internal/service/`.
- Property-based tests verifying subtask progress rollup calculations and recursive tree hierarchies.
- Deterministic time injection using mock clocks to test due date transitions (`overdue`, `due-today`, `due-soon`).

### 3.2 Storage In-Memory SQLite Testing
- Tests in `internal/storage/` execute against an in-memory SQLite instance (`file::memory:?cache=shared`).
- Migrations run automatically during test suite setup to verify schema migrations up and down.

### 3.3 Synthetic Bubble Tea `tea.Msg` Pump Testing
- TUI models are tested headlessly without requiring an active terminal emulator.
- Synthetic message streams (`tea.KeyMsg`, `tea.WindowSizeMsg`, custom domain messages) are passed directly to `model.Update(msg)`.
- Assertions verify state transitions, returned `tea.Cmd` behaviors, and `model.View()` string outputs.

### 3.4 CLI Golden File & Integration Testing
- Cobra CLI commands are tested via standard `bytes.Buffer` capture for `stdout` and `stderr`.
- Golden file testing (`testdata/*.golden`) validates exact CLI tabular layouts and `--json` schemas.

---

## 4. Performance & Operational Benchmarks

| Metric | Target | Verification Method |
| :--- | :--- | :--- |
| **CLI Cold Start (`tusk list`)** | < 15 ms | `hyperfine 'bin/tusk list'` |
| **CLI Help / Version (`tusk -h`)** | < 5 ms | Zero database connection touch |
| **Binary Size** | < 25 MB | Single statically compiled executable |
| **TUI Frame Time** | < 16 ms (60 FPS) | Pure `View()` rendering with no dynamic memory allocations |
| **Test Suite Execution** | < 5 seconds | Parallelized Go unit and integration tests |
