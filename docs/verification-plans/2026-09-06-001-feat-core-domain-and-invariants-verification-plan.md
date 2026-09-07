---
feature-id: 2026-09-06-001-feat-core-domain-and-invariants
plan-source: docs/plans/2026-09-06-001-feat-core-domain-and-invariants-plan.md
surface-profiles: [library-core-domain]
status: Verified - Passed
evidence-scope: Verified local execution
---

# Feature 001: Core Domain & Invariants Verification Plan

## 1. Verification Contract

- **Behavior Under Test**: Pure Go domain logic, strongly typed value objects (`Status`, `Priority`, `Tag`), task state machine lifecycle with timestamp tracking, integer mathematical progress rollup, cycle-free recursive tree validation, and deterministic filtering/sorting.
- **Public Contracts**:
  - `internal/core/errors.go`: 15 domain error sentinels (including `ErrInvalidProgress`).
  - `internal/core/status.go`: `ParseStatus`, `Status.CanTransitionTo`.
  - `internal/core/priority.go`: `ParsePriority`, `Priority.Weight`.
  - `internal/core/tag.go`: `NormalizeTag`, `NormalizeTags`, `NormalizeTagSlice`.
  - `internal/core/task.go`: `NewTask`, `Task.TransitionTo`, `Task.Update`, `Task.SetParent`, `Task.SetProgress`, `Task.SetRollupProgress`, `Task.Clone`.
  - `internal/core/rollup.go`: `CalculateProgress(task Task, subtasks []Task) int`.
  - `internal/core/tree.go`: `BuildTree`, `DetectCycles(taskID string, proposedParentID *string, ...) error`, `ValidateHierarchyDepth(taskSubtreeDepth int, proposedParentID string, ...) error`.
  - `internal/core/filter.go`: `FilterTasks(tasks []Task, filter TaskFilter) []Task`, `SortTasks(tasks []Task, order []SortOrder)`. `TaskFilter` with `RootOnly bool`, `SortByID` in `SortField`.
- **Supported Environments**: Go 1.24+ runtime on Linux, macOS, and Windows.
- **Local Evidence Available**: Pure Go table-driven unit tests, property tests, race condition detection, and micro-benchmarks.
- **Hosted/Manual Evidence Required**: None. Feature 001 is hermetic and pure.
- **Explicit Non-Goals**: Database persistence testing, SQLite schema migrations, CLI flag parsing, or TUI rendering.

---

## 2. Requirement Coverage

| Requirement | Unit | Scenario IDs | Evidence Tier |
| :--- | :--- | :--- | :--- |
| **Error Sentinels** | Unit 001-1 | CORE-ERR-N1, CORE-ERR-B1 | Focused Unit |
| **Status State Machine** | Unit 001-2 | CORE-STS-N1, CORE-STS-B1, CORE-STS-F1 | Focused Unit |
| **Priority Weights** | Unit 001-2 | CORE-PRI-N1, CORE-PRI-B1 | Focused Unit |
| **Tag Normalization** | Unit 001-2 | CORE-TAG-N1, CORE-TAG-B1 | Focused Unit |
| **Task Lifecycle & Dates** | Unit 001-3 | CORE-TSK-N1, CORE-TSK-B1, CORE-TSK-R1 | Focused Unit |
| **Rollup Arithmetic** | Unit 001-4 | CORE-ROL-N1, CORE-ROL-B1, CORE-ROL-P1 | Focused Property / Table |
| **Tree & Cycle Invariants**| Unit 001-5 | CORE-TRE-N1, CORE-TRE-B1, CORE-TRE-F1, CORE-TRE-BM1 | Focused Graph / Benchmark |
| **Filtering & Sorting** | Unit 001-6 | CORE-FLT-N1, CORE-FLT-B1, CORE-FLT-C1 | Focused Unit |
| **Aggregate Domain Suite** | All | CORE-AGG-ALL | Aggregate `make validate` (domain coverage threshold $\ge 95\%$ target; 98.0% measured) |

---

## 3. Scenarios

### Error Taxonomy
- [x] **CORE-ERR-N1 Normal Path**: All 15 error sentinels (`ErrTaskNotFound`, `ErrCyclicDependency`, `ErrInvalidProgress`, `ErrInvalidTaskID`, `ErrDuplicateTaskID`, `ErrInvalidDepth`, `ErrTraversalLimitExceeded`, etc.) have distinct error string representations and non-nil identity.
- [x] **CORE-ERR-B1 Boundary**: Wrapped errors (`fmt.Errorf("context: %w", ErrEmptyTitle)`) correctly unwrap and match via `errors.Is`.

### Value Objects (`Status`, `Priority`, `Tag`)
- [x] **CORE-STS-N1 Normal Path**: `ParseStatus` converts `"todo"`, `"in-progress"`, `"blocked"`, `"done"` to valid enums.
- [x] **CORE-STS-B1 Boundary**: `ParseStatus` trims whitespace and handles case insensitivity (`"TODO"` -> `StatusTodo`).
- [x] **CORE-STS-F1 Injected Failure**: Invalid string `"review"` returns `ErrInvalidStatus`. Invalid transition `done -> blocked` returns `ErrInvalidStatusTransition`. Reopening `done -> in-progress` succeeds.
- [x] **CORE-PRI-N1 Normal Path**: `ParsePriority` converts `"urgent"`, `"high"`, `"medium"`, `"low"` and integers 1–4 to `Priority` enums.
- [x] **CORE-PRI-B1 Boundary**: Invalid priority strings or integers outside 1–4 return `ErrInvalidPriority`.
- [x] **CORE-TAG-N1 Normal Path**: `NormalizeTag` converts `"#Backend"` to `Tag("backend")`. `NormalizeTags` and `NormalizeTagSlice` deduplicate duplicate tags in slice, sort lexicographically, and return an empty non-nil slice for nil or empty inputs.
- [x] **CORE-TAG-B1 Boundary**: Tag containing spaces or invalid symbols returns `ErrInvalidTag`. Tags exceeding 32 characters are rejected.

### Task Entity & State Transitions
- [x] **CORE-TSK-N1 Normal Path**: `NewTask` creates valid entity with `StatusTodo`, `Progress = 0`, and non-zero `CreatedAt`/`UpdatedAt`.
- [x] **CORE-TSK-B1 Boundary**: Empty title returns `ErrEmptyTitle`. Title of 256 characters returns `ErrTitleTooLong`. `SetProgress` validates $0 \le \text{progress} \le 99$ for non-done tasks (100 strictly on done), rejecting out-of-bounds with `ErrInvalidProgress`. `SetRollupProgress` enforces exact match against `CalculateProgress` when subtasks are supplied, and requires complete subtasks when setting 100 on a non-done parent.
- [x] **CORE-TSK-C1 Defensive Copying**: `Task.Clone` returns fully independent pointer and slice fields (`ParentID`, `DueDate`, `CompletedAt`, `Tags`), with nil-vs-empty slice preservation and reflection-based field-exhaustiveness enforcement.
- [x] **CORE-TSK-R1 Recovery / Reopen**: Moving status from `StatusDone` to `StatusInProgress` sets `CompletedAt = nil` and updates `UpdatedAt`. `SetParent` updates `ParentID` and `UpdatedAt`, rejecting self-parenting with `ErrSelfParenting`.

### Mathematical Progress Rollup
- [x] **CORE-ROL-N1 Normal Path**: Leaf task preserves explicitly assigned manual progress ($0 \le \text{progress} \le 99$) or evaluates to $100\%$ when `StatusDone`.
- [x] **CORE-ROL-B1 Boundary**: Integer floor rounding: subtasks with progresses [100, 0, 0] calculate parent progress as $\lfloor \frac{100}{3} \rfloor = 33\%$.
- [x] **CORE-ROL-P1 Property Invariant**: If all direct subtasks have `status == StatusDone`, rollup progress is strictly $100\%$ regardless of individual child progress integers.
- [x] **CORE-ROL-P2 Nested Hierarchy Invariant**: Intermediate non-done parent reporting rolled-up 100% progress preserves its full 100% contribution to grandparent rollup calculations without degradation to 99%.

### Tree Traversal, Hierarchy & Cycle Prevention
- [x] **CORE-TRE-N1 Normal Path**: `BuildTree` assembles a list of flat tasks into a forest of `TaskNode` trees with correct `Depth` attributes.
- [x] **CORE-TRE-B1 Boundary**: Root promotion with `nil` `proposedParentID` returns `nil` without invoking lookup. Subtree reparenting validates `parentDepth + 1 + taskSubtreeDepth <= MaxHierarchyDepth`, prioritizing cycle detection over depth limits for cycles $\ge 11$ nodes (returns `ErrCyclicDependency`).
- [x] **CORE-TRE-F1 Cycle & Orphan Injection**: In an existing tree `A -> B -> C`, attempting to set `A.ParentID = &C` fails cycle check with `ErrCyclicDependency`. `BuildTree` returns `ErrTaskNotFound` if parent ID is missing from slice.
- [x] **CORE-TRE-BM1 Benchmark**: `DetectCycles` traversal on a 10-level hierarchy completes in $< 500\text{ns}$ per check (observed 301ns/op, 0 allocs).

### Filtering & Sorting
- [x] **CORE-FLT-N1 Normal Path**: `FilterTasks` accurately filters a list by status slice, priority slice, tag intersection, title substring, and `RootOnly` flag.
- [x] **CORE-FLT-B1 Boundary & Sorting**: Multi-key sort with `SortByDueDate` places `nil` due dates last on `SortAsc` (first on `SortDesc`), with `SortByID ASC` deterministic final tie-breaker.
- [x] **CORE-FLT-C1 Zero-Match Edge Case**: Queries matching zero tasks return empty non-nil slices.

---

## 4. Commands and Environments

| Tier | Command | Environment | Planned Evidence |
| :--- | :--- | :--- | :--- |
| **Focused Unit** | `go test -v ./internal/core/...` | Local Linux/Darwin/Win | All unit tests pass in $< 1\text{s}$ |
| **Race Detector** | `go test -race -v ./internal/core/...` | Local Linux/Darwin | Zero data race conditions detected |
| **Tree Traversal Benchmark** | `go test -bench=BenchmarkTreeTraversal -benchmem ./internal/core/...` | Local Linux | Sub-microsecond execution (< 500ns, 0 allocs; observed 301ns/op) |
| **Tree Build Benchmark** | `go test -bench=BenchmarkBuildTree -benchmem ./internal/core/...` | Local Linux | Scalable forest allocation (< 1µs/task, < 5 allocs/task; observed ~38µs/100 tasks, 481 allocs) |
| **Aggregate Gate**| `make validate && go test -cover -race ./internal/core/...` | Local | Strict format, vet, unit tests, race checks, and $\ge 95\%$ domain coverage target |

---

## 5. Fixtures and Test Doubles

- **Fixtures**: Table-driven task matrices with predetermined depths, cycle topologies, and progress distributions.
- **Test Doubles**: Pure in-memory parent lookup functions (`func(id string) (*string, error)`) to test cycle detection without external storage.
- **Deterministic Clocks**: Explicit injection of `now time.Time` in `NewTask` and `TransitionTo` to guarantee deterministic timestamp assertions.

---

## 6. Execution Record

*(This section will be populated with execution timestamps, commit SHAs, and test outputs during the implementation phase.)*

| Date | Commit SHA | Environment | Command | Result | Evidence Ref |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 2026-09-06 | `3eeedf4` | Go 1.24 Linux x86_64 | `make validate` | Pass (0 race, 0 vet) | `docs/workorders/2026-09-06-001-feat-core-domain-and-invariants-issues-workorder.md` |
| 2026-09-06 | `9ab407c` | Go 1.24 Linux x86_64 | `go test -cover -race ./internal/core/...` | Pass (98.0% coverage) | `docs/workorders/2026-09-06-001-feat-core-domain-and-invariants-issues-workorder.md` |
| 2026-09-06 | `9ab407c` | Go 1.24 Linux x86_64 | `go test -bench=BenchmarkTreeTraversal -benchmem ./internal/core/...` | Pass (301ns/op, 0 allocs) | `docs/workorders/2026-09-06-001-feat-core-domain-and-invariants-issues-workorder.md` |
| 2026-09-07 | `b248bb2` | Go 1.24 Linux x86_64 | `go test -bench=BenchmarkBuildTree -benchmem ./internal/core/...` | Pass (37.8µs/100 tasks, 481 allocs) | `docs/workorders/2026-09-06-001-feat-core-domain-and-invariants-issues-workorder.md` |
