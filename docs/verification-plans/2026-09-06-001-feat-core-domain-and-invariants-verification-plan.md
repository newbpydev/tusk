---
feature-id: 2026-09-06-001-feat-core-domain-and-invariants
plan-source: docs/plans/2026-09-06-001-feat-core-domain-and-invariants-plan.md
surface-profiles: [library-core-domain]
status: Draft - not executed
evidence-scope: Planning only
---

# Feature 001: Core Domain & Invariants Verification Plan

## 1. Verification Contract

- **Behavior Under Test**: Pure Go domain logic, strongly typed value objects (`Status`, `Priority`, `Tag`), task state machine lifecycle with timestamp tracking, integer mathematical progress rollup, cycle-free recursive tree validation, and deterministic filtering/sorting.
- **Public Contracts**:
  - `internal/core/errors.go`: 10 domain error sentinels.
  - `internal/core/status.go`: `ParseStatus`, `Status.CanTransitionTo`.
  - `internal/core/priority.go`: `ParsePriority`, `Priority.Weight`.
  - `internal/core/tag.go`: `NormalizeTag`, `NormalizeTags`.
  - `internal/core/task.go`: `NewTask`, `Task.TransitionTo`, `Task.Update`, `Task.SetProgress`.
  - `internal/core/rollup.go`: `CalculateProgress([]Task) int`.
  - `internal/core/tree.go`: `BuildTree`, `DetectCycles`, `ValidateHierarchyDepth`.
  - `internal/core/filter.go`: `FilterTasks`, `SortTasks`.
- **Supported Environments**: Go 1.24+ runtime on Linux, macOS, and Windows.
- **Local Evidence Available**: Pure Go table-driven unit tests, property tests, race condition detection, and micro-benchmarks.
- **Hosted/Manual Evidence Required**: None. Feature 001 is hermetic and pure.
- **Explicit Non-Goals**: Database persistence testing, SQLite schema migrations, CLI flag parsing, or TUI rendering.

---

## 2. Requirement Coverage

| Requirement | Unit | Scenario IDs | Evidence Tier |
| :--- | :--- | :--- | :--- |
| **Error Sentinels** | Unit 001-1 | ERR-N1, ERR-B1 | Focused Unit |
| **Status State Machine** | Unit 001-2 | STS-N1, STS-B1, STS-F1 | Focused Unit |
| **Priority Weights** | Unit 001-2 | PRI-N1, PRI-B1 | Focused Unit |
| **Tag Normalization** | Unit 001-2 | TAG-N1, TAG-B1 | Focused Unit |
| **Task Lifecycle & Dates** | Unit 001-3 | TSK-N1, TSK-B1, TSK-R1 | Focused Unit |
| **Rollup Arithmetic** | Unit 001-4 | ROL-N1, ROL-B1, ROL-P1 | Focused Property / Table |
| **Tree & Cycle Invariants**| Unit 001-5 | TRE-N1, TRE-B1, TRE-F1, TRE-BM1 | Focused Graph / Benchmark |
| **Filtering & Sorting** | Unit 001-6 | FLT-N1, FLT-B1, FLT-C1 | Focused Unit |

---

## 3. Scenarios

### Error Taxonomy
- [ ] **ERR-N1 Normal Path**: All error sentinels (`ErrTaskNotFound`, `ErrCyclicDependency`, etc.) have distinct error string representations and satisfy `errors.Is(err, sentinel)`.
- [ ] **ERR-B1 Boundary**: Wrapped errors (`fmt.Errorf("context: %w", ErrEmptyTitle)`) correctly unwrap and match via `errors.Is`.

### Value Objects (`Status`, `Priority`, `Tag`)
- [ ] **STS-N1 Normal Path**: `ParseStatus` converts `"todo"`, `"in-progress"`, `"blocked"`, `"done"` to valid enums.
- [ ] **STS-B1 Boundary**: `ParseStatus` trims whitespace and handles case insensitivity (`"TODO"` -> `StatusTodo`).
- [ ] **STS-F1 Injected Failure**: Invalid string `"review"` returns `ErrInvalidStatus`. Invalid transition `done -> blocked` returns `ErrInvalidStatusTransition`.
- [ ] **PRI-N1 Normal Path**: `ParsePriority` converts `"urgent"`, `"high"`, `"medium"`, `"low"` and integers 1–4 to `Priority` enums.
- [ ] **PRI-B1 Boundary**: Invalid priority strings or integers outside 1–4 return `ErrInvalidPriority`.
- [ ] **TAG-N1 Normal Path**: `NormalizeTag` converts `"#Backend"` to `Tag("backend")`. Deduplicates duplicate tags in slice.
- [ ] **TAG-B1 Boundary**: Tag containing spaces or invalid symbols returns `ErrInvalidTag`. Tags exceeding 32 characters are rejected.

### Task Entity & State Transitions
- [ ] **TSK-N1 Normal Path**: `NewTask` creates valid entity with `StatusTodo`, `Progress = 0`, and non-zero `CreatedAt`/`UpdatedAt`.
- [ ] **TSK-B1 Boundary**: Empty title returns `ErrEmptyTitle`. Title of 256 characters returns `ErrTitleTooLong`. Title of exactly 255 characters passes.
- [ ] **TSK-R1 Recovery / Reopen**: Moving status from `StatusDone` to `StatusInProgress` sets `CompletedAt = nil` and updates `UpdatedAt`.

### Mathematical Progress Rollup
- [ ] **ROL-N1 Normal Path**: Subtasks with progresses [100, 50, 0] calculate parent progress as $\lfloor \frac{150}{3} \rfloor = 50\%$.
- [ ] **ROL-B1 Boundary**: Integer floor rounding: subtasks with progresses [100, 0, 0] calculate parent progress as $\lfloor \frac{100}{3} \rfloor = 33\%$.
- [ ] **ROL-P1 Property Invariant**: If all direct subtasks have `status == StatusDone`, rollup progress is strictly $100\%$ regardless of individual child progress integers.

### Tree Traversal, Hierarchy & Cycle Prevention
- [ ] **TRE-N1 Normal Path**: `BuildTree` assembles a list of flat tasks into a forest of `TaskNode` trees with correct `Depth` attributes.
- [ ] **TRE-B1 Self-Parenting Boundary**: Attempting to set a task's parent to its own ID returns `ErrSelfParenting`.
- [ ] **TRE-F1 Cycle Injection**: In an existing tree `A -> B -> C`, attempting to set `A.ParentID = &C` fails cycle check with `ErrCyclicDependency`.
- [ ] **TRE-BM1 Benchmark**: `DetectCycles` traversal on a 10-level hierarchy completes in $< 500\text{ns}$ per check.

### Filtering & Sorting
- [ ] **FLT-N1 Normal Path**: `FilterTasks` accurately filters a list by status slice, priority slice, tag intersection, and title substring.
- [ ] **FLT-B1 Boundary**: Empty filter returns original task slice intact.
- [ ] **FLT-C1 Deterministic Sorting**: Multi-key sort (Priority DESC, DueDate ASC, CreatedAt ASC) produces identical order across 100 consecutive runs.

---

## 4. Commands and Environments

| Tier | Command | Environment | Planned Evidence |
| :--- | :--- | :--- | :--- |
| **Focused Unit** | `go test -v ./internal/core/...` | Local Linux/Darwin/Win | All unit tests pass in $< 1\text{s}$ |
| **Race Detector** | `go test -race -v ./internal/core/...` | Local Linux/Darwin | Zero data race conditions detected |
| **Benchmarks** | `go test -bench=. -benchmem ./internal/core/...` | Local Linux | Sub-microsecond execution, $< 5$ allocations |
| **Aggregate Gate**| `make validate` | Local | Strict format, vet, unit tests, and race checks |

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
| *Planned* | *Pending* | Go 1.24 Linux x86_64 | `go test -v ./internal/core/...` | *Pending* | `docs/workorders/2026-09-06-001-feat-core-domain-and-invariants-issues-workorder.md` |
