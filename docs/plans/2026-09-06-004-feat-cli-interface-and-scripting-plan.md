---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: 2026-09-06-001-feat-tusk-modern-task-system-plan.md
---

# Feature Plan 004: CLI Interface & Scripting

## Goal Capsule

Implement the Unix-composable Command Line Interface (CLI) in `internal/cli/` using Cobra. Provide human-friendly, ANSI-styled tabular outputs for terminal developers, machine-readable `--json` output for shell scripts and AI agents, strict exit code guarantees, and lazy database initialization ensuring instant startup for `--help` and `--version`.

---

## Technical Design & Scope

### 1. Command Tree (`internal/cli/`)
- Root command: `tusk [subcommand] [flags]`
- Subcommands:
  - `add <title>`: Flag support for priority (`-p`), due date (`-d`), tags (`-t`), parent (`--parent`), notes (`-n`).
  - `list`: Filters by status (`-s`), priority (`-p`), tags (`-t`), due date range (`--due`).
  - `done <id>`: Marks specified task as complete.
  - `edit <id>`: Updates attributes of an existing task.
  - `delete <id>`: Removes task with `--recursive` / `--force` flags.
  - `tree [id]`: Visual hierarchy rendering.
  - `stats`: Metrics summary (completion rate, overdue counts, velocity).
  - `tui`: Launches the interactive terminal UI.

### 2. Output Formatting & Composability
- **Tabular Formatter**: Uses Lipgloss for border rendering, status badges (green `DONE`, yellow `TODO`, blue `IN-PROGRESS`, red `BLOCKED`), and dynamic column truncation based on terminal width.
- **JSON Formatter**: Emitted whenever `--json` is passed. Clean, unformatted or indented JSON written strictly to `stdout`. All logs/notices routed to `stderr`.

### 3. Exit Code Contract
- `0`: Success.
- `1`: Domain or runtime failure (task not found, validation error, database error).
- `2`: Command syntax error or unknown flag.

---

## Verification Scenarios

1. Golden file tests for tabular `tusk list` output matching expected terminal formatting.
2. JSON schema validation tests asserting `tusk list --json` matches `TaskDTO` JSON schema.
3. Cold start benchmark verifying `tusk --help` runs in $< 5\text{ms}$ with no database access.
4. Exit code test verifying correct codes (0, 1, 2) across valid and invalid CLI executions.
