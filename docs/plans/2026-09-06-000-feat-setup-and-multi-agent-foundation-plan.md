---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: 2026-09-06-001-feat-tusk-modern-task-system-plan.md
---

# Feature Plan 000: Setup & Multi-Agent Foundation

## Goal Capsule

Establish the foundational infrastructure for the Tusk reboot: a clean Go 1.24+ module scaffold, a canonical Makefile governing all quality gates, shell scripts with TDD test suites, unified multi-agent steering via `AGENTS.md` and IDE rules, and domain taxonomy defined in `CONCEPTS.md`.

---

## Technical Design & Scope

### 1. Build & Validation Infrastructure
- **Makefile**: Single-command orchestrator exposing `setup`, `fmt`, `vet`, `test`, `test-unit`, `race`, `validate`, `build`, and `clean`.
- **Scripts**:
  - `scripts/setup.sh`: Validates Go toolchain (Go 1.24+) and prepares environment.
  - `scripts/fmt.sh`: Enforces strict `gofmt -s -w` formatting.
  - `scripts/test/test_scripts.sh`: Self-testing test harness asserting that scripts behave correctly on valid and erroneous conditions.

### 2. Multi-Agent Steering Architecture
- **Canonical Root**: `AGENTS.md` specifying persona, first-principles protocol, TDD rules, prohibited anti-patterns, and project commands.
- **IDE Configurations**:
  - `.cursor/rules/*.mdc`: Modular rules for Cursor IDE covering TDD mandate, Go idioms, Bubble Tea TUI constraints, and storage/sqlc.
  - `.claude/` & `.omp/`: Symlink or mirrored configuration pointing to `AGENTS.md`.
  - Compatibility with Opencode and Kilo via repository-root `AGENTS.md`.
- **Domain Taxonomy**: `CONCEPTS.md` defining project vocabulary (Task Tree, Root Task, Subtask, Rollup Progress, Display Order, Storage Engine, Active Panel).

### 3. Go Project Scaffold
- `go.mod`: Module `github.com/newbpydev/tusk`, Go 1.24.
- Minimal compilable `cmd/tusk/main.go` and placeholder test ensuring `make validate` exits 0 cleanly.

---

## Verification Scenarios

1. `make setup` runs without error and detects Go toolchain.
2. `make fmt` runs cleanly and formats Go source code.
3. `./scripts/test/test_scripts.sh` passes all script regression checks.
4. `make validate` executes `fmt`, `vet`, `test`, and `race` with exit code 0.
5. `make build` produces a working binary at `bin/tusk`.
