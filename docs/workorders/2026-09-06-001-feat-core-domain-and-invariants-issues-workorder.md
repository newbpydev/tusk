---
feature-id: 2026-09-06-001-feat-core-domain-and-invariants
plan-source: docs/plans/2026-09-06-001-feat-core-domain-and-invariants-plan.md
verification-plan: docs/verification-plans/2026-09-06-001-feat-core-domain-and-invariants-verification-plan.md
status: Closed - Verified
evidence-scope: Verified local execution
---

# Feature 001: Core Domain & Invariants Issue Workorder

## 1. Issue Register

| ID | Source | Owner / Lens | Severity | Status | Impact | Next Action | Retest / Evidence |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **CORE-ISS-001** | Deepening Audit | Correctness / State Machine | P1 | Fixed in Plan | Legacy allowed invalid state jumps (`done -> blocked`). | Defined explicit `CanTransitionTo` matrix in plan. | Unit 001-2 test suite (`TestStatusTransitions`). |
| **CORE-ISS-002** | Deepening Audit | Performance / Graph Theory | P2 | Fixed in Plan | Legacy had no cycle detection on subtask reparenting, risking infinite UI loops. | Specified `DetectCycles` with depth limits in `internal/core/tree.go`. | Unit 001-5 test suite (`TestDetectCycles`). |
| **CORE-ISS-003** | Deepening Audit | Data Integrity / Rollup Math | P2 | Fixed in Plan | Ambiguity in progress integer rounding when division has fractional remainder. | Defined integer floor policy ($\lfloor \frac{\sum P}{N} \rfloor$). | Unit 001-4 test suite (`TestCalculateProgress`). |
| **CORE-ISS-004** | Deepening Audit | Domain Boundaries / Coupling | P1 | Fixed in Plan | Legacy mixed natural language date parsing inside domain entities. | Clarified that natural date parsing belongs to Feature 003 service layer; `Task` accepts pure `time.Time`. | Architecture boundary review in Unit 001-3. |
| **CORE-ISS-005** | Doc Review | Contract / Rollup Math | P1 | Fixed in Plan | `CalculateProgress` signature lacked task context & wiped out leaf manual progress. | Updated signature to `CalculateProgress(task Task, subtasks []Task) int`; leaf preserves manual progress (0–99). | Unit 001-4 tests (`TestCalculateProgress_Leaf`, `CORE-ROL-N1`). |
| **CORE-ISS-006** | Doc Review | API Ambiguity / Filtering | P1 | Fixed in Plan | `TaskFilter.ParentID *string` could not distinguish root tasks from unconstrained queries. | Added `RootOnly bool` to `TaskFilter` with explicit matching semantics. | Unit 001-6 tests (`TestFilterTasks`, `CORE-FLT-N1`). |
| **CORE-ISS-007** | Doc Review | Error Taxonomy / Validation | P2 | Fixed in Plan | Split progress setters lacked typed sentinel error, exact rollup match contract, and leaf reset. | Added `ErrInvalidProgress` sentinel, `SetProgress` (requires valid status including zero-value rejection, 0–99 on non-done, 100 on done, leaf-only at service boundary), `SetRollupProgress` (rejects invalid parent/child status and corrupt child progress, exact match with subtasks rollup on non-done), `TransitionTo` idempotent lifecycle repair, and `ResetLeaf`. | Unit 001-3 tests (`TestTask_SetProgress_Validation`, `TestTask_SetRollupProgress`, `TestTask_ResetLeaf`, `TestTask_TransitionTo_IdempotentRepair`, `TestTask_ZeroStatusRejected`, `CORE-ERR-N1`, `CORE-TSK-B1`). |
| **CORE-ISS-008** | Doc Review | Graph Invariants / Reparenting | P1 | Fixed in Plan | `ValidateHierarchyDepth` only traversed upward, ignoring descendant subtree depth during reparenting. | Updated signature to accept `taskSubtreeDepth int` and validate `parentDepth + 1 + subtreeDepth <= MaxHierarchyDepth`. | Unit 001-5 tests (`TestValidateHierarchyDepth_Subtree`, `CORE-TRE-B1`). |
| **CORE-ISS-009** | Doc Review | API Ergonomics / Graph | P2 | Fixed in Plan | `DetectCycles` required string parameter, preventing callers passing `nil` on root promotion. | Changed `proposedParentID` to `*string`; `nil` returns `nil` immediately. | Unit 001-5 tests (`TestDetectCycles_RootPromotion`, `CORE-TRE-B1`). |
| **CORE-ISS-010** | Doc Review | Error Contracts / Tree | P1 | Fixed in Plan | `BuildTree` error conditions on orphaned parents and cyclic loops were undefined. | Specified `BuildTree` returns `ErrTaskNotFound` on missing parents, `ErrCyclicDependency` on loops. | Unit 001-5 tests (`TestBuildTree_Errors`, `CORE-TRE-F1`). |
| **CORE-ISS-011** | Doc Review | Determinism / Sorting | P2 | Fixed in Plan | `SortTasks` with `SortByDueDate` lacked nulls-last ordering and deterministic tie-breaker. | Specified nulls-last on `SortAsc` and implicit `SortByID ASC` final tie-breaker. | Unit 001-6 tests (`TestSortTasks_MultiKey`, `CORE-FLT-B1`). |
| **CORE-ISS-012** | Doc Review | Lifecycle / Entity Mutation | P2 | Fixed in Plan | `Task` entity omitted dedicated `SetParent` method, risking stale `UpdatedAt` on reparenting. | Added `func (t *Task) SetParent(parentID *string, now time.Time) error` with `ErrSelfParenting` check. | Unit 001-3 tests (`TestTask_SetParent`, `CORE-TSK-R1`). |
| **CORE-ISS-013** | Doc Review | Architecture / Query Boundaries | P2 | Fixed in Plan | In-memory filtering and sorting created apparent duplication with SQLite storage queries. | Delineated operational boundary: in-memory for TUI live search; primary persistence queries in storage repository. | Architecture boundary review in Section 2.7. |
| **CORE-ISS-014** | Code Review | Error Taxonomy / Hierarchy Guards | P1 | Verified | Missing explicit domain sentinels for empty task IDs, duplicate IDs in trees, and negative subtree depths. | Added `ErrInvalidTaskID`, `ErrDuplicateTaskID`, and `ErrInvalidDepth` (14 total sentinels). | Unit tests in `errors_test.go`, `task_test.go`, and `tree_test.go`. |
| **CORE-ISS-015** | Code Review | Defensive Copying / Pointers | P1 | Verified | Value copies in tree/filter paths aliased pointers and slices. | Added `Task.Clone() Task` with nil-preservation and reflection-based field-exhaustiveness test. | Unit tests in `task_test.go` and `filter_test.go` (`CORE-TSK-C1`). |
| **CORE-ISS-016** | Code Review | Value Objects / Triplet | P2 | Verified | Value-object helpers and trim-prefix-trim order lacked explicit contracts in planning triplet. | Documented `NormalizeTagSlice`, `Tag.String`, `Status.IsValid`, `Status.IsTerminal`, `Priority.String`, and `Priority.IsValid` with tests. | Unit tests in `tag_test.go`, `status_test.go`, and `priority_test.go` (`CORE-TAG-N1`, `CORE-STS-N1`, `CORE-PRI-N1`). |
---

## 2. Issue Details

### CORE-ISS-001: Explicit Status State Machine & Reopening Invariants
- **Phase Found**: Planning / Deepening Audit
- **Owner / Review Lens**: Correctness & Reliability Lens
- **Severity**: P1
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-2 (`internal/core/status.go`)
- **Planning Gap**: In the legacy codebase, task statuses were arbitrary strings without validation. Tasks could jump from `done` to `blocked` without properly resetting timestamps.
- **Decision & Fix**: Defined strict enum `Status` with `CanTransitionTo` validation method. Reopening a `done` task explicitly mandates setting `CompletedAt = nil`.
- **Retest / Closure Evidence**: Unit test `TestStatusTransitions` asserting `done -> blocked` returns `ErrInvalidStatusTransition`.

### CORE-ISS-002: Recursive Subtask Cycle Prevention
- **Phase Found**: Planning / Deepening Audit
- **Owner / Review Lens**: Plan Architecture & Algorithmic Correctness
- **Severity**: P2
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-5 (`internal/core/tree.go`)
- **Planning Gap**: Legacy Tusk allowed assigning any task ID as a parent, creating potential circular dependencies (`A -> B -> A`), causing recursive algorithms to stack overflow.
- **Decision & Fix**: Added `DetectCycles` ancestor traversal algorithm and `MaxHierarchyDepth = 10` ceiling, returning `ErrCyclicDependency` or `ErrMaxDepthExceeded`.
- **Retest / Closure Evidence**: Unit test `TestDetectCycles` verifying 1-node, 2-node, and 5-node loops fail.

### CORE-ISS-003: Deterministic Integer Floor for Progress Rollup
- **Phase Found**: Planning / Deepening Audit
- **Owner / Review Lens**: Mathematical Precision Lens
- **Severity**: P2
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-4 (`internal/core/rollup.go`)
- **Planning Gap**: Fractional percentages (e.g. 1 task done out of 3 = 33.333%) could produce rounding drift or floating point inconsistency.
- **Decision & Fix**: Mandated strict integer floor arithmetic: $\lfloor \frac{\sum P}{N} \rfloor$. Added explicit invariant: if all children are `StatusDone`, progress is strictly $100\%$.
- **Retest / Closure Evidence**: Table-driven unit test `TestCalculateProgress` covering division boundaries.

### CORE-ISS-004: Decoupling Date Parsing from Pure Domain
- **Phase Found**: Planning / Deepening Audit
- **Owner / Review Lens**: Simplicity & Layering Boundary Lens
- **Severity**: P1
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-3 (`internal/core/task.go`)
- **Planning Gap**: Initial drafts considered putting natural language date parsing inside `internal/core/task.go`, adding unnecessary string parsing complexity to pure domain models.
- **Decision & Fix**: `internal/core/` only knows `time.Time`. Natural language date parsing (`today`, `tomorrow`, `+2d`) is isolated in `internal/service/dateparse/` (Feature 003).
- **Retest / Closure Evidence**: Verified `internal/core` has zero regex/time-parsing dependencies outside standard `time.Time`.

### CORE-ISS-005: CalculateProgress Leaf Task Contract & Signature
- **Phase Found**: Planning / Document Review
- **Owner / Review Lens**: Coherence, Feasibility, Adversarial, Product-Lens, Whole-Doc
- **Severity**: P1
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-4 (`internal/core/rollup.go`)
- **Planning Gap**: `CalculateProgress(subtasks []Task)` did not take the target task, making Rule 1 (`if len(subtasks) == 0: 100 if t.Status == StatusDone else 0`) unimplementable, and overwrote manual progress on leaf tasks.
- **Decision & Fix**: Updated signature to `func CalculateProgress(task Task, subtasks []Task) int`. Rule 1 updated to preserve explicitly assigned manual progress ($0 \le \text{progress} \le 99$) on non-done leaf tasks.
- **Retest / Closure Evidence**: Unit test `TestCalculateProgress_Leaf` and scenario `CORE-ROL-N1`.

### CORE-ISS-006: TaskFilter Root Task Disambiguation
- **Phase Found**: Planning / Document Review
- **Owner / Review Lens**: Feasibility & Coherence Lenses
- **Severity**: P1
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-6 (`internal/core/filter.go`)
- **Planning Gap**: `TaskFilter.ParentID *string` defaulting to `nil` made it impossible to distinguish between filtering for root tasks (`parent_id IS NULL`) and not filtering by parentage at all.
- **Decision & Fix**: Added `RootOnly bool` to `TaskFilter`. Documented that `RootOnly: true` filters for roots (`ParentID == nil`), while `!RootOnly && ParentID == nil` leaves parentage unconstrained.
- **Retest / Closure Evidence**: Unit test `TestFilterTasks` and scenario `CORE-FLT-N1`.

### CORE-ISS-007: Out-of-Bounds Progress Sentinel
- **Phase Found**: Planning / Document Review
- **Owner / Review Lens**: Feasibility Lens
- **Severity**: P2
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-1 & Unit 001-3 (`internal/core/errors.go`, `internal/core/task.go`)
- **Planning Gap**: `Task.SetProgress` lacked a typed domain sentinel error and clear separation between manual leaf progress updates and service-layer subtask rollup updates.
- **Decision & Fix**: Added immutable `ErrInvalidProgress = Error("task progress must be an integer between 0 and 100")`. Specified `SetProgress` for manual leaf updates (valid status enum required via `ErrInvalidStatus`, including zero-value rejection; $0 \le \text{progress} \le 99$ on non-done tasks, $100$ strictly on done tasks, leaf-only precondition owned by the service boundary) and `SetRollupProgress` for service-layer rollup calculations ($0 \le \text{progress} \le 100$, rejecting parents and subtasks with invalid status enums via `ErrInvalidStatus` and corrupt child progress, enforcing exact match with `CalculateProgress(*t, subtasks)` on non-done tasks with subtasks), `TransitionTo` lifecycle repair on idempotent calls, and `ResetLeaf` under a non-restorative deletion policy to assign leaf progress when subtasks are empty (`len(subtasks) == 0`).
- **Retest / Closure Evidence**: Unit tests `TestTask_SetProgress_Validation`, `TestTask_SetRollupProgress`, `TestTask_ResetLeaf`, and scenarios `CORE-ERR-N1`, `CORE-TSK-B1`, `CORE-ROL-P2`.

### CORE-ISS-008: ValidateHierarchyDepth Subtree Reparenting
- **Phase Found**: Planning / Document Review
- **Owner / Review Lens**: Adversarial & Feasibility Lenses
- **Severity**: P1
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-5 (`internal/core/tree.go`)
- **Planning Gap**: Upward traversal from `proposedParentID` failed to inspect existing descendant depth below the moving task, allowing reparented subtrees to silently exceed `MaxHierarchyDepth`.
- **Decision & Fix**: Updated signature to `func ValidateHierarchyDepth(taskSubtreeDepth int, proposedParentID string, lookupParent func(id string) (*string, error)) error` checking `parentDepth + 1 + taskSubtreeDepth <= MaxHierarchyDepth`.
- **Retest / Closure Evidence**: Unit test `TestValidateHierarchyDepth_Subtree` and scenario `CORE-TRE-B1`.

### CORE-ISS-009: DetectCycles Root Promotion Pointer Parameter
- **Phase Found**: Planning / Document Review
- **Owner / Review Lens**: Adversarial Lens
- **Severity**: P2
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-5 (`internal/core/tree.go`)
- **Planning Gap**: `DetectCycles` accepted `proposedParentID string`, preventing callers promoting tasks to root from passing `nil` and causing spurious lookup errors on empty strings.
- **Decision & Fix**: Changed signature to `func DetectCycles(taskID string, proposedParentID *string, lookupParent func(id string) (*string, error)) error`. When `proposedParentID == nil`, returns `nil` immediately.
- **Retest / Closure Evidence**: Unit test `TestDetectCycles_RootPromotion` and scenario `CORE-TRE-B1`.

### CORE-ISS-010: BuildTree Error Conditions & Orphan Handling
- **Phase Found**: Planning / Document Review
- **Owner / Review Lens**: Feasibility & Whole-Doc Lenses
- **Severity**: P1
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-5 (`internal/core/tree.go`)
- **Planning Gap**: `BuildTree` declared an `error` return without specifying failure modes for orphaned nodes or circular references.
- **Decision & Fix**: Specified that `BuildTree` returns `ErrTaskNotFound` if non-root tasks reference missing parent IDs, and `ErrCyclicDependency` if cyclic loops are detected.
- **Retest / Closure Evidence**: Unit test `TestBuildTree_Errors` and scenario `CORE-TRE-F1`.

### CORE-ISS-011: SortTasks Nulls-Last & Deterministic Tie-Breaker
- **Phase Found**: Planning / Document Review
- **Owner / Review Lens**: Feasibility & Whole-Doc Lenses
- **Severity**: P2
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-6 (`internal/core/filter.go`)
- **Planning Gap**: `SortByDueDate` lacked null pointer comparison rules, risking runtime panics, and lacked a deterministic final tie-breaker on identical sort keys.
- **Decision & Fix**: Defined that `nil` due dates sort last on `SortAsc` (first on `SortDesc`), and `SortTasks` always appends `SortByID ASC` as an implicit deterministic final tie-breaker.
- **Retest / Closure Evidence**: Unit test `TestSortTasks_MultiKey` and scenario `CORE-FLT-B1`.

### CORE-ISS-012: Task.SetParent Method and Lifecycle
- **Phase Found**: Planning / Document Review
- **Owner / Review Lens**: Coherence Lens
- **Severity**: P2
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-3 (`internal/core/task.go`)
- **Planning Gap**: `Task` lacked a managed `SetParent` method, risking stale `UpdatedAt` timestamps and bypassing `ErrSelfParenting` validation during task reparenting.
- **Decision & Fix**: Added `func (t *Task) SetParent(parentID *string, now time.Time) error` to Section 2.4, updating `UpdatedAt = now` and validating against self-parenting.
- **Retest / Closure Evidence**: Unit test `TestTask_SetParent` and scenario `CORE-TSK-R1`.

### CORE-ISS-013: Delineation of In-Memory Filtering vs Storage Queries
- **Phase Found**: Planning / Document Review
- **Owner / Review Lens**: Adversarial & Product-Lens Lenses
- **Severity**: P2
- **Status**: Fixed in Plan
- **Affected Requirement / Unit**: Unit 001-6 (`internal/core/filter.go`)
- **Planning Gap**: In-memory `FilterTasks` and `SortTasks` raised questions about architectural duplication with SQLite `sqlc` queries.
- **Decision & Fix**: Delineated operational boundary: `core.FilterTasks` serves client-side in-memory live search in the Bubble Tea TUI without database round-trips; storage queries remain in `ports.TaskRepository`.
- **Retest / Closure Evidence**: Architectural boundary review in Section 2.7.

### CORE-ISS-014: Extension of Domain Error Taxonomy to 14 Sentinels
- **Phase Found**: Code Review Audit
- **Owner / Review Lens**: Correctness & Reliability Lens
- **Severity**: P1
- **Status**: Closed - Verified
- **Affected Requirement / Unit**: Unit 001-1 (`internal/core/errors.go`), Unit 001-3 (`task.go`), Unit 001-5 (`tree.go`)
- **Defect**: Tree and task validation lacked explicit typed sentinels for empty IDs, duplicate task IDs in trees, negative subtree depths, and traversal-step exhaustion, causing silent misreporting or masking defects as generic cycles.
- **Decision & Fix**: Defined and tested dedicated sentinels: `ErrInvalidTaskID`, `ErrDuplicateTaskID`, and `ErrInvalidDepth`, expanding the taxonomy to 14 sentinels while detecting cycles of arbitrary depth without traversal caps.
- **Retest / Closure Evidence**: Unit tests in `errors_test.go`, `task_test.go`, and `tree_test.go` verifying clean sentinel typing and matching.
---

## 3. Review-Lens Sign-Offs

| Lens | Status | Evidence / Findings | Retest Required |
| :--- | :--- | :--- | :--- |
| **Plan Architecture** | Completed | 6 sequential implementation units; strict acyclic dependencies; zero external package imports. | None |
| **Product / Scope** | Completed | 100% traceable to Brainstorm Product Contract (`2026-09-06-001-feat-tusk-modern-task-system-plan.md`). | None |
| **Correctness & Reliability** | Completed | Explicit state transitions, cycle prevention, floor arithmetic, and sentinel errors resolved in planning. | None |
| **Test Strategy** | Completed | TDD Red-First tests planned for all 6 units; synthetic fixtures, benchmarks, and race detector gates defined. | None |
| **Performance & Concurrency**| Completed | In-memory operations operate in $< 1\mu\text{s}$; pure value copies eliminate cross-goroutine mutation risks. | None |
| **Simplicity & Maintainability**| Completed | Standard library only; no reflection, no external dependencies, no complex generic acrobatics. | None |

---

## 4. Implementation Release Gate

*Note: All items remain unchecked during planning. Checkboxes will be marked exclusively during execution with verified command outputs.*

- [x] Planned units implemented (Units 001-1 through 001-6).
- [x] Focused unit tests pass (`make test`).
- [x] Race detector checks pass (`make race`).
- [x] Benchmarks pass with expected performance (`make bench-tree` observed 314ns/op, 0 allocs; `make bench-build` observed 73.6µs/op, 484 allocs across 100 tasks).
- [x] Domain test coverage satisfies $\ge 95\%$ mandate (`make coverage` observed 98.2% under `-race`).
- [x] Aggregate repository validation passes (`make validate`).
- [x] All issues fixed (CORE-ISS-001 through CORE-ISS-016).
- [x] Remaining unaccepted issues: 0.
