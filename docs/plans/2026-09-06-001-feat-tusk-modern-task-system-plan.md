---
artifact_contract: ce-unified-plan/v1
artifact_readiness: requirements-only
product_contract_source: ce-brainstorm
---

# Feature Plan: Tusk Modern Task Management System

## Goal Capsule

Reboot Tusk into a blazing-fast, zero-friction, keyboard-first task management system tailored for software engineers. Tusk delivers instant out-of-the-box local operation with zero background daemons, zero external database configurations, and zero interactive authentication barriers. It provides both a Unix-composable CLI optimized for terminal scripting and AI agent tool integration, and an interactive Bubble Tea terminal user interface (TUI) enforcing strict unidirectional Elm architecture with zero layout shifts or visual flicker.

---

## Product Contract

### 1. Scope & Problem Framing

Developers require task management that lives inside the terminal and executes with sub-millisecond responsiveness. Legacy solutions suffer from:
1. Heavy operational overhead (mandatory Docker daemons, external PostgreSQL servers).
2. Clunky CLI ergonomics (interactive password prompts blocking headless execution, failure on `--help` if databases are unreachable).
3. Brittle terminal interfaces (state mutation inside render passes, terminal flickering, unhandled window resizes).

Tusk 2026 solves this by separating concerns cleanly: an embedded local SQLite engine by default, an expressive domain core managing recursive subtask hierarchies with automatic mathematical progress rollup, and a decoupled presentation tier delivering both a scriptable CLI and a rich TUI.

---

### 2. Core Entities & Taxonomy

```
                   +-------------------+
                   |     Root Task     |
                   +---------+---------+
                             |
                +------------+------------+
                |                         |
        +-------v-------+         +-------v-------+
        | Child Subtask |         | Child Subtask |
        +-------+-------+         +---------------+
                |
        +-------v-------+
        | Leaf Subtask  |
        +---------------+
```

#### 2.1 Task Entity
- `id`: ULID or UUID string (lexicographically sortable, collision-free).
- `title`: String (1-255 characters, required, trimmed).
- `description`: Optional Markdown/text notes.
- `status`: Strongly typed enum: `todo`, `in-progress`, `blocked`, `done`.
- `priority`: Strongly typed enum: `low` (1), `medium` (2), `high` (3), `urgent` (4).
- `parent_id`: Nullable ID referencing parent task. Null designates a Root Task.
- `progress`: Integer percentage (0–100): manually assigned on leaf tasks (0–99, default 0% on creation), automatically calculated from subtasks on parent tasks, $100\%$ when `status == done`.
- `tags`: Set of normalized lowercase alphanumeric strings (`#backend`, `#bug`).
- `due_date`: Nullable UTC timestamp parsed from natural language input.
- `created_at`: UTC timestamp set on entity creation.
- `updated_at`: UTC timestamp updated on any mutation.
- `completed_at`: Nullable UTC timestamp set when status becomes `done`, cleared if reopened.

#### 2.2 Task Hierarchy & Rollup Engine
- **Recursive Depth**: Arbitrary n-level parent-child tree.
- **Acyclic Enforcement**: Cycle detection on creation/move; an ancestor can never become its own descendant.
- **Progress Calculation**:
  - Leaf Task: explicitly assigned manual progress ($0\%$–$99\%$) if `status != done` (fresh tasks default $0\%$), $100\%$ if `status == done`.
  - Parent Task: $\lfloor \frac{1}{N} \sum_{i=1}^N \text{subtask}_i.\text{progress} \rfloor$.
  - State Sync: When all child subtasks are marked `done`, the parent status can optionally auto-transition to `done`. If any subtask is reopened, a `done` parent reverts to `in-progress`.

---

### 3. Storage & Runtime Contract

#### 3.1 Embedded SQLite Engine
- **Default Storage Path**: Conforms to XDG specification: `~/.local/share/tusk/tusk.db`. Can be overridden via `TUSK_DB_PATH`.
- **Driver**: Pure Go CGO-free driver (`modernc.org/sqlite` or `ncruces/go-sqlite3`).
- **Pragmas**:
  - `PRAGMA journal_mode = WAL;` (Concurrent readers, non-blocking writes).
  - `PRAGMA synchronous = NORMAL;` (High durability with minimal I/O stalls).
  - `PRAGMA foreign_keys = ON;` (Strict relational integrity).
  - `PRAGMA busy_timeout = 5000;` (Graceful concurrency lock handling).
- **Embedded Migrations**: Schema migrations are embedded into the Go binary and executed idempotently on first access.
- **Pluggable Architecture**: The repository layer is accessed via `ports.TaskRepository`, enabling future optional backend drivers (e.g., PostgreSQL) without domain changes.

---

### 4. CLI Interface Contract

#### 4.1 Invocation & Exit Codes
- Fast Startup: `tusk --help` and `tusk --version` must execute in $< 5\text{ms}$ with zero database connection overhead.
- Exit Codes:
  - `0`: Successful execution.
  - `1`: Operational error (task not found, validation failure, storage failure).
  - `2`: Syntax or command-line flag error.

#### 4.2 Subcommand Grammar
- `tusk add "<title>" [flags]`: Create a new task.
  - Flags: `-p, --priority`, `-d, --due`, `-t, --tags`, `--parent <id>`, `-n, --notes`.
- `tusk list [flags]`: List tasks.
  - Flags: `-s, --status`, `-p, --priority`, `-t, --tags`, `--due <date>`, `--all`, `--json`.
- `tusk done <id>`: Mark a task and its subtasks as done.
- `tusk edit <id> [flags]`: Modify title, priority, status, due date, or tags.
- `tusk delete <id> [--recursive]`: Remove a task (requires confirmation or `--force` if it has subtasks).
- `tusk tree [id]`: Display task hierarchy in a visual ASCII/Unicode tree.
- `tusk stats`: Display completion metrics, overdue task count, and productivity summaries.
- `tusk tui`: Launch the interactive terminal UI.

#### 4.3 Machine-Readable Output (`--json`)
Every query command (`list`, `tree`, `stats`, `add`, `done`) must support `--json`. When active:
- Standard output (`stdout`) contains only valid JSON.
- Progress bars, color codes, and human-facing messages are suppressed or emitted to `stderr`.

---

### 5. Interactive Terminal UI (TUI) Contract

#### 5.1 Architecture & State Flow
- Built with Charm's Bubble Tea v1.3+, Lipgloss v1.1+, and Bubbles v0.21+.
- **Strict Unidirectional Elm Pattern**:
  - `View() string` is a **pure function**. It never mutates model fields, never initializes managers, and never allocates styles dynamically.
  - All mutations occur exclusively in `Update(tea.Msg) (tea.Model, tea.Cmd)`.
  - Asynchronous queries and disk I/O are handled via `tea.Cmd`.

#### 5.2 Multi-Panel Layout & Viewports
- **Left Panel (Task Tree / List)**: Collapsible task groups (`Today`, `Upcoming`, `Backlog`, `Completed`), visual selection cursor, status badges.
- **Right Panel (Task Details & Notes)**: Full description rendered with Markdown formatting, subtask progress bar, due date, tags, and timeline history.
- **Bottom Panel (Status & Keymap Bar)**: Context-aware keybinding hints (`j/k` navigate, `x` done, `a` add, `e` edit, `q` quit, `?` help).
- **Zero Layout Shifts**: Fixed borders and flexbox-style width/height calculations guarantee layout stability across any terminal window $> 80\times24$.

#### 5.3 Keyboard Navigation Contract
- Vim bindings: `h/j/k/l` for movement, `g/G` for top/bottom.
- Action triggers:
  - `a`: Open modal form to add task.
  - `x` or `Space`: Toggle task completion.
  - `e`: Open modal form to edit selected task.
  - `d` / `dd`: Delete task with confirmation prompt.
  - `/`: Focus search input bar with live debounce filtering.
  - `Tab` / `Shift+Tab`: Cycle active focus across panels.
  - `Esc`: Cancel modal or clear search.

---

### 6. Non-Goals & Prohibitions

1. **No External Authentication for Local CLI**: Single-user local installations must never require username/password logins or network identity tokens.
2. **No Monolithic Global Database Pools**: Global connection variables (`var Pool *...`) are forbidden; connection handles must be injected cleanly through constructors.
3. **No Uncoordinated Goroutine Pools in Domain Logic**: Background worker pools with `sync.Map` caches are prohibited. Bubble Tea commands manage concurrency in the UI, while standard Go concurrency patterns manage CLI executions.
4. **No Side-Effects in `View()`**: Any function called within `View()` must be read-only and idempotent.

---

### 7. Acceptance Criteria & Verification Matrix

| Area | Criterion | Verification Method |
| :--- | :--- | :--- |
| **CLI Startup** | `bin/tusk --help` runs without database initialized | Automated execution without `tusk.db` |
| **Task Lifecycle** | Create, view, complete, reopen task preserves metadata | In-memory SQLite regression test |
| **Subtask Rollup** | Parent progress accurately reflects subtask completions | Pure mathematical domain unit test |
| **Acyclic Invariant** | Making a parent a child of its own subtask returns error | Negative test case in domain suite |
| **JSON Composability** | `tusk list --json` outputs valid parseable JSON to stdout | Automated JSON schema validation |
| **TUI Purity** | Calling `View()` 100 times consecutively produces identical output without mutating model | Headless Bubble Tea unit test |
