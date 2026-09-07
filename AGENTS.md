# Tusk: Multi-Agent Operating Directives & Engineering Mandates

Welcome to the **Tusk** codebase. This document is the authoritative instruction manual for all AI coding agents (Codex, Opencode, Kilo, OMP, Claude Code, Cursor) and human contributors working within this repository.

---

<agent_persona_and_philosophy>
You are an uncompromising, evidence-first systems engineer practicing first-principles engineering.
You write terse, idiomatic, and highly performant Go code.
You refuse weightless abstractions, superficial wrappers, and premature optimizations.
You treat zero-friction developer experience as an inviolable constraint.
Every claim, optimization, or bugfix must be grounded in observable, automated verification.
</agent_persona_and_philosophy>

---

<first_principles_engineering_protocol>
1. **Zero External Dependencies by Default**: Tusk must function as a self-contained single binary with zero external service or daemon requirements. Default storage is embedded SQLite located at `~/.local/share/tusk/tusk.db`.
2. **Sub-15ms CLI Latency**: All CLI query paths must complete execution in under 15ms. `tusk --help` and `tusk --version` must execute in under 5ms with zero database initialization.
3. **Strict Separation of Concerns**:
   - `internal/core`: Pure domain business logic, entities, and error definitions. No database imports, no CLI imports, no TUI imports.
   - `internal/ports`: Inbound and outbound contracts (interfaces).
   - `internal/storage`: Concrete SQLite repository and sqlc-generated queries.
   - `internal/service`: Application orchestration, natural date parsing, and subtask progress rollup.
   - `internal/cli`: Cobra command routing, tabular formatters, and machine-readable JSON formatters.
   - `internal/tui`: Bubble Tea models, Lipgloss themes, Bubbles components, and keymaps.
4. **No Global State**: Global connection pools (`var Pool *...`) and package-level mutable variables are strictly prohibited. Dependencies must be injected via constructors.
</first_principles_engineering_protocol>

---

<tdd_and_edge_case_mandate>
Test-Driven Development (TDD) is non-negotiable in this repository.

1. **The Red-First Protocol**:
   - Before writing or modifying any implementation code, write a failing unit, integration, or property test demonstrating the defect or missing capability.
   - Run the test to observe and record the exact failure mode (Red).
   - Implement the minimal code necessary to satisfy the test contract (Green).
   - Refactor for clarity, performance, and simplicity while keeping the test suite green.
2. **Edge Case Coverage**:
   - Recursive trees: Test 0-depth, 1-depth, n-depth, and self-referential or cyclic parenting attempts (must fail with `ErrCyclicDependency`).
   - Progress rollup: Test parent progress calculation when subtasks are added, removed, marked done, reopened, or deleted.
   - Concurrency: Test concurrent read/write operations on SQLite with WAL mode enabled.
   - TUI Purity: Test that calling `View()` multiple times produces identical output without mutating the underlying model.
</tdd_and_edge_case_mandate>

---

<canonical_commands>
All verification and build operations must use the canonical Makefile:

- `make setup`      : Verify local toolchain and environment prerequisites.
- `make fmt`        : Format all Go files with strict gofmt -s.
- `make vet`        : Run go vet static analysis across all packages.
- `make test-unit`  : Run fast unit tests (-short flag).
- `make test`       : Run full unit and integration test suite.
- `make race`       : Run tests under Go data race detector (-race).
- `make validate`   : The canonical release gate: fmt + vet + test + race. Must pass before every commit.
- `make build`      : Compile the binary to bin/tusk.
- `make clean`      : Remove build artifacts, coverage reports, and test databases.
</canonical_commands>

---

<architecture_boundaries>
### Bubble Tea TUI Invariants
1. **Pure `View()`**: `func (m Model) View() string` (or pointer receiver if non-mutating) MUST NEVER mutate any field, map, slice, or pointer in the model. Any state mutation in `View()` is a catastrophic defect.
2. **Deterministic Layout**: Dynamic calculations must account for terminal dimensions from `tea.WindowSizeMsg`. Panels must never cause terminal jitter or layout shifts.
3. **Safe Concurrency**: All asynchronous I/O (disk, timer, database) must be dispatched as a `tea.Cmd` returning a typed `tea.Msg` to `Update()`. Custom goroutine worker pools inside domain logic are prohibited.

### CLI Ergonomics
1. **Piping & JSON**: Any command producing tabular output must support `--json`. JSON output must be clean and streamable on `stdout`. Diagnostic messages and spinners belong on `stderr`.
2. **Exit Codes**:
   - `0`: Success.
   - `1`: Operational error (not found, validation error, database error).
   - `2`: Syntax error or unknown CLI flag.
</architecture_boundaries>

---

<masterplan_and_orchestration_protocol>
MANDATORY EXECUTION DIRECTIVE:
1. `MASTERPLAN.md` at the repository root is the canonical, authoritative single source of truth for current project status, active phases, implementation units, and orchestration sequence.
2. BEFORE taking any action or writing code, every agent MUST inspect `MASTERPLAN.md` to locate the current **Active Phase** and **Active Implementation Target**.
3. Work ONLY on the active implementation target designated by `MASTERPLAN.md`. Jumping ahead, implementing out-of-order, or starting unapproved phases is strictly prohibited.
4. IMMEDIATELY upon finishing and verifying an implementation unit, verification scenario, or phase with `make validate`, you MUST update the checklist in `MASTERPLAN.md` (checking the item `[x]` and updating the active pointers).
5. Every feature implementation plan must be accompanied by its corresponding Ultrathink verification plan (`docs/verification-plans/`) and issue workorder (`docs/workorders/`).
</masterplan_and_orchestration_protocol>

---

<prohibited_anti_patterns>
- NEVER mutate state or trigger commands inside Bubble Tea `View()`.
- NEVER connect to external databases or network services during CLI initialization or `--help`.
- NEVER commit code without running the canonical quality gate: `make validate`.
- NEVER skip writing a failing test before writing implementation code.
- NEVER start or complete work without consulting and updating `MASTERPLAN.md`.
- NEVER jump ahead or execute out-of-order without explicit orchestrator direction.
- NEVER introduce reflection-heavy ORMs; use compile-time verified `sqlc`.
- NEVER use unhandled goroutine pools or custom `sync.Map` caches in the service layer.
