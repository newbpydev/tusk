---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: 2026-09-06-001-feat-tusk-modern-task-system-plan.md
feature-id: 2026-09-06-001-feat-core-domain-and-invariants
title: Core Domain Entities, Hierarchies & Business Invariants
---

# Feature Plan 001: Core Domain Entities, Hierarchies & Business Invariants

## 1. Scope Capsule

### Goal & Operator Value
Provide the pure Go business logic and domain core for Tusk in `internal/core/`. This package establishes the inviolable rules of the task system: strongly typed states, priority weights, tag normalization, recursive subtask trees, cycle prevention, and mathematical progress rollups. By keeping this package 100% pure (zero database, zero network, zero CLI, zero UI dependencies), the system guarantees deterministic, microsecond-latency evaluation and unit test isolation.

### Actors & Affected Systems
- **Direct Consumers**: `internal/ports/`, `internal/service/`, `internal/storage/`.
- **Indirect Consumers**: `internal/cli/`, `internal/tui/`.
- **Isolation Guarantee**: `internal/core/` does not import any other package within `tusk/` and depends solely on the Go standard library (`time`, `fmt`, `strings`, `errors`, `regexp`, `unicode`).

### In Scope
- Strongly typed domain enums: `Status` (`todo`, `in-progress`, `blocked`, `done`) and `Priority` (`urgent`, `high`, `medium`, `low`).
- Entity definition: `Task` with ULID/UUID identifier, metadata, timestamps, and parent linkage.
- Tag value object with normalization rules (lowercase, alphanumeric + dashes, trimmed, deduped).
- Validated state transition machine with automatic `CompletedAt` lifecycle management.
- Mathematical subtask progress rollup calculation with deterministic integer floor arithmetic.
- Recursive tree model, acyclic cycle detection (`ErrCyclicDependency`), and defensive hierarchy depth ceiling (`ErrMaxDepthExceeded`, default 10 levels) preventing recursive stack overflow and terminal indentation clipping.
- Task filtering query criteria and deterministic multi-key sorting rules.
- Strongly typed domain sentinel errors.

### Out of Scope
- Persistence mechanisms, SQL queries, or SQLite drivers (owned by Feature 002).
- Natural language date string parsing (owned by Feature 003 service layer).
- Terminal rendering, Lipgloss styling, or Bubble Tea models (owned by Feature 005).
- CLI flag parsing and Cobra command definitions (owned by Feature 004).

### Surface Profiles
- **Library / Core Domain**: Pure Go exports, deterministic algorithms, thread-safe value copies, zero allocations on hot paths, $\ge 95\%$ domain test coverage target (98.0% measured).

### Evidence Boundary
- **Local Evidence Only**: Pure Go unit tests (`go test -v ./internal/core/...`), race detection (`go test -race ./internal/core/...`), property-based assertions, and micro-benchmarks (`go test -bench=. ./internal/core/...`). No external infrastructure, daemons, or network access required.

---

## 2. Technical Design & Exact Public Contracts

### 2.1 Domain Errors (`internal/core/errors.go`)
```go
package core

type Error string

func (e Error) Error() string

type SelfParentingError string

func (e SelfParentingError) Error() string
func (e SelfParentingError) Is(target error) bool

const (
	ErrTaskNotFound            = Error("task not found")
	ErrEmptyTitle              = Error("task title cannot be empty")
	ErrTitleTooLong            = Error("task title exceeds maximum length of 255 characters")
	ErrInvalidStatus           = Error("invalid task status")
	ErrInvalidPriority         = Error("invalid task priority")
	ErrInvalidStatusTransition = Error("invalid status transition")
	ErrCyclicDependency        = Error("cyclic dependency detected: a task cannot be its own ancestor")
	ErrSelfParenting           = SelfParentingError("task cannot reference itself as parent")
	ErrMaxDepthExceeded        = Error("maximum subtask hierarchy depth exceeded")
	ErrInvalidTag              = Error("invalid tag format: tags must be alphanumeric with hyphens")
	ErrInvalidProgress         = Error("task progress must be an integer between 0 and 100")
	ErrInvalidTaskID           = Error("invalid task id: id cannot be empty")
	ErrDuplicateTaskID         = Error("duplicate task id in hierarchy")
	ErrInvalidDepth            = Error("invalid hierarchy depth: depth cannot be negative")
	ErrTraversalLimitExceeded  = Error("hierarchy traversal limit exceeded")
)
```

### 2.2 Strongly Typed Enums (`internal/core/status.go` & `internal/core/priority.go`)

#### Status Enum & State Machine
```go
type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusBlocked    Status = "blocked"
	StatusDone       Status = "done"
)

func ParseStatus(s string) (Status, error)
func (s Status) IsValid() bool
func (s Status) IsTerminal() bool // true for StatusDone
func (s Status) CanTransitionTo(next Status) bool
```
**Transition Matrix**:
- From `todo`: Can transition to `in-progress`, `blocked`, `done`.
- From `in-progress`: Can transition to `todo`, `blocked`, `done`.
- From `blocked`: Can transition to `todo`, `in-progress`, `done`.
- From `done`: Can transition to `todo`, `in-progress` (reopening). Transition to `blocked` directly is invalid.

#### Priority Enum
```go
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
	PriorityUrgent Priority = 4
)

func ParsePriority(s string) (Priority, error)
func (p Priority) String() string
func (p Priority) Weight() int
func (p Priority) IsValid() bool
```

### 2.3 Tag Value Object (`internal/core/tag.go`)
```go
type Tag string

func NormalizeTag(raw string) (Tag, error)
func NormalizeTags(raw []string) ([]Tag, error)
func NormalizeTagSlice(raw []Tag) ([]Tag, error)
func (t Tag) String() string
```
**Normalization Rules**:
1. Strip leading `#` if present.
2. Trim whitespace.
3. Convert to lowercase ASCII.
4. Validate regex: `^[a-z0-9]+(-[a-z0-9]+)*$` (length 1–32 chars).
5. Deduplicate and sort lexicographically.

### 2.4 Task Entity (`internal/core/task.go`)
```go
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      Status     `json:"status"`
	Priority    Priority   `json:"priority"`
	ParentID    *string    `json:"parent_id,omitempty"`
	Progress    int        `json:"progress"` // 0 - 100
	Tags        []Tag      `json:"tags,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type NewTaskParams struct {
	ID          string
	Title       string
	Description string
	Priority    Priority
	ParentID    *string
	Tags        []string
	DueDate     *time.Time
	Now         time.Time
}

func NewTask(params NewTaskParams) (*Task, error)
func (t *Task) TransitionTo(next Status, now time.Time) error
func (t *Task) Update(title, desc string, priority Priority, tags []Tag, dueDate *time.Time, now time.Time) error
func (t *Task) SetParent(parentID *string, now time.Time) error
func (t *Task) SetProgress(progress int, now time.Time) error
func (t *Task) SetRollupProgress(progress int, subtasks []Task, now time.Time) error
func (t *Task) IsRoot() bool
func (t *Task) IsDone() bool
func (t Task) Clone() Task
```
**Entity Mutation Contracts**:
- `SetParent` reassigns `ParentID` and updates `UpdatedAt = now`. Returns `ErrSelfParenting` if `parentID != nil && *parentID == t.ID`.
- `SetProgress` sets manual leaf progress ($0 \le \text{progress} \le 99$ for non-done tasks, strictly $100$ for done tasks), returning `ErrInvalidProgress` on out-of-bounds inputs or when attempting to set 100 on a non-done task, and updates `UpdatedAt = now`.
- `SetRollupProgress` sets progress computed by the rollup engine ($0 \le \text{progress} \le 100$). When subtasks are provided, progress must match `CalculateProgress(*t, subtasks)`, returning `ErrInvalidProgress` on value mismatches. Setting 100 on a non-done parent requires subtasks to be provided and all complete.
- `Clone` returns a deep copy of `Task` with independent pointer and slice fields (`ParentID`, `DueDate`, `CompletedAt`, `Tags`).

### 2.5 Progress Rollup Engine (`internal/core/rollup.go`)
```go
func CalculateProgress(task Task, subtasks []Task) int
```
**Mathematical Rules**:
1. If `len(subtasks) == 0`:
   $$\text{Progress} = \begin{cases} 100 & \text{if } task.\text{Status} == \text{StatusDone} \\ task.\text{Progress} & \text{otherwise (preserves assigned manual progress, 0--99)} \end{cases}$$
2. If `len(subtasks) > 0`:
   $$\text{Progress} = \left\lfloor \frac{1}{N} \sum_{i=1}^N \text{subtask}_i.\text{Progress} \right\rfloor$$
   Clamped between $0$ and $100$.
3. If all subtasks are `StatusDone`, rollup progress is guaranteed to be $100\%$.
4. Reopening any subtask recalculates the parent's progress proportionally.

### 2.6 Tree Hierarchy & Cycle Detection (`internal/core/tree.go`)
```go
const MaxHierarchyDepth = 10

type TaskNode struct {
	Task     Task        `json:"task"`
	Children []*TaskNode `json:"children,omitempty"`
	Depth    int         `json:"depth"`
}

func BuildTree(tasks []Task) ([]*TaskNode, error)
func DetectCycles(taskID string, proposedParentID *string, lookupParent func(id string) (*string, error)) error
func ValidateHierarchyDepth(taskSubtreeDepth int, proposedParentID string, lookupParent func(id string) (*string, error)) error
```
**Hierarchy & Tree Contracts**:
- `MaxHierarchyDepth = 10`: While Tusk guarantees arbitrary n-level recursive trees conceptually, the domain core enforces a defensive ceiling of 10 levels to protect against runaway recursion, stack overflow, and visual terminal line truncation during tree indentation. Cycles are strictly detected via `DetectCycles` regardless of depth.
- `DetectCycles`: When `proposedParentID == nil`, returns `nil` immediately (root tasks have no parent and cannot introduce cycles).
- `ValidateHierarchyDepth`: `taskSubtreeDepth` is the maximum depth of existing descendants below the moving task (0 for leaf tasks). The check validates that `parentDepth + 1 + taskSubtreeDepth <= MaxHierarchyDepth`.
- `BuildTree`: Returns `ErrTaskNotFound` if any non-root task's `ParentID` references an ID absent from the slice, and `ErrCyclicDependency` if unrooted loops or cycles are detected within the slice.

### 2.7 Filter & Sorting Models (`internal/core/filter.go`)
```go
type SortField string

const (
	SortByID        SortField = "id"
	SortByPriority  SortField = "priority"
	SortByDueDate   SortField = "due_date"
	SortByCreatedAt SortField = "created_at"
	SortByTitle     SortField = "title"
)

type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

type SortOrder struct {
	Field     SortField
	Direction SortDirection
}

type TaskFilter struct {
	Statuses   []Status
	Priorities []Priority
	Tags       []Tag
	ParentID   *string
	RootOnly   bool
	DueBefore  *time.Time
	DueAfter   *time.Time
	SearchTerm string
}

func FilterTasks(tasks []Task, filter TaskFilter) []Task
func SortTasks(tasks []Task, order []SortOrder)
```
**Filter & Sorting Contracts**:
- **Operational Boundary**: `FilterTasks` and `SortTasks` operate in-memory on task slices, specifically providing instant sub-millisecond filtering and sorting for interactive Bubble Tea TUI views (live search while typing) and CLI tree transformations without issuing database round-trips. Primary persistence queries and persistent filtering remain the responsibility of `ports.TaskRepository` via SQLite `sqlc`.
- `FilterTasks`: When `RootOnly` is true, matches root tasks (`ParentID == nil`). When `RootOnly` is false and `ParentID != nil`, matches that specific parent. When `!RootOnly && ParentID == nil`, parentage filter is unconstrained (matches any task).
- `SortTasks`: `SortByDueDate` places tasks with `nil` `DueDate` last when `Direction` is `SortAsc` (and first when `SortDesc`). `SortTasks` always appends `SortOrder{Field: SortByID, Direction: SortAsc}` as an implicit deterministic final tie-breaker if not already specified.

---

## 3. Implementation Units & TDD Sequencing

```mermaid
graph TD
    U1[Unit 001-1: Errors & Sentinels] --> U2[Unit 001-2: Value Objects Status, Priority, Tag]
    U2 --> U3[Unit 001-3: Task Entity & State Machine]
    U3 --> U4[Unit 001-4: Mathematical Progress Rollup]
    U3 --> U5[Unit 001-5: Tree Traversal & Cycle Detection]
    U4 --> U6[Unit 001-6: Task Filtering & Sorting Engine]
    U5 --> U6
```

### Unit 001-1: Domain Sentinels and Error Taxonomy
- **Goal**: Define domain error sentinels.
- **Files**:
  - `internal/core/errors.go`
  - `internal/core/errors_test.go`
- **Dependencies**: None.
- **Red Test**: `TestErrorsExist`: assert `errors.Is(ErrTaskNotFound, ErrTaskNotFound)`, ensure all sentinels are unique instances and have non-empty error strings.
- **Verification**: `go test -v -run TestErrors ./internal/core/...`
- **Review Lenses**: Correctness, simplicity.

### Unit 001-2: Strongly Typed Value Objects (`Status`, `Priority`, `Tag`)
- **Goal**: Implement enums, parsing, serialization, and normalization logic.
- **Files**:
  - `internal/core/status.go`, `internal/core/status_test.go`
  - `internal/core/priority.go`, `internal/core/priority_test.go`
  - `internal/core/tag.go`, `internal/core/tag_test.go`
- **Dependencies**: Unit 001-1.
- **Red Tests**:
  - `TestParseStatus`: Valid status parsing, invalid status returns `ErrInvalidStatus`.
  - `TestStatusTransitions`: Valid transitions pass; `done -> blocked` returns `ErrInvalidStatusTransition`.
  - `TestParsePriority`: Maps 1-4 and strings ("urgent", "high", "medium", "low"); rejects 0 or 5 with `ErrInvalidPriority`.
  - `TestNormalizeTag`: Strips `#`, converts `Backend` -> `backend`, rejects spaces/special chars with `ErrInvalidTag`.
- **Verification**: `go test -v -run "TestParse|TestNormalize|TestStatus" ./internal/core/...`
- **Review Lenses**: API contract, correctness.

### Unit 001-3: Task Entity & State Machine
- **Goal**: Implement `Task` entity, constructor `NewTask`, and mutation methods with timestamp lifecycle.
- **Files**:
  - `internal/core/task.go`
  - `internal/core/task_test.go`
- **Dependencies**: Unit 001-2.
- **Red Tests**:
  - `TestNewTask_Validation`: Reject empty title (`ErrEmptyTitle`), title > 255 chars (`ErrTitleTooLong`).
  - `TestTask_TransitionToDone`: Transitioning to `StatusDone` sets `CompletedAt` to `now`.
  - `TestTask_Reopen`: Transitioning from `StatusDone` to `StatusInProgress` clears `CompletedAt` (sets to `nil`).
  - `TestTask_SetParent`: Reassigning parent updates `ParentID` and `UpdatedAt`; self-parenting returns `ErrSelfParenting`.
  - `TestTask_SetProgress_Validation`: Valid percentages update `Progress` and `UpdatedAt`; values < 0 or > 100 return `ErrInvalidProgress`; setting 100 on non-done task returns `ErrInvalidProgress`.
  - `TestTask_SetRollupProgress`: Rollup-calculated progress allows 100 on non-done parent; values < 0 or > 100 return `ErrInvalidProgress`.
Verification: `go test -v -run TestTask ./internal/core/...`
- **Review Lenses**: Data integrity, state machine correctness.

### Unit 001-4: Mathematical Progress Rollup Engine
- **Goal**: Implement `CalculateProgress` with floor integer math and edge case handling.
- **Files**:
  - `internal/core/rollup.go`
  - `internal/core/rollup_test.go`
- **Dependencies**: Unit 001-3.
  - `TestCalculateProgress_Leaf`: 0% when `todo`, preserves manual progress (e.g. 50% when in-progress), 100% when `done`.
  - `TestCalculateProgress_Subtasks`: Parent task with 3 subtasks (100%, 50%, 0%) -> average is 50%.
  - `TestCalculateProgress_FloorRounding`: Parent task with 3 subtasks (100%, 0%, 0%) -> 33% (integer floor).
  - `TestCalculateProgress_AllDone`: All subtasks done -> strictly 100%.
  - `TestCalculateProgress_NestedHierarchy100`: Intermediate parent with rolled-up 100% progress preserves 100% contribution to grandparent rollup without degradation.
- **Verification**: `go test -v -run TestCalculateProgress ./internal/core/...`
- **Review Lenses**: Mathematical accuracy, boundaries.

### Unit 001-5: Hierarchical Tree Traversal and Cycle Detection
- **Goal**: Implement `DetectCycles`, `ValidateHierarchyDepth`, and `BuildTree`.
- **Files**:
  - `internal/core/tree.go`
  - `internal/core/tree_test.go`
- **Dependencies**: Unit 001-3.
  - `TestDetectCycles_DirectSelf`: Setting `taskA.parent = taskA` returns `ErrSelfParenting`.
  - `TestDetectCycles_RootPromotion`: Passing `nil` proposedParentID returns `nil` without calling lookupParent.
  - `TestDetectCycles_TwoNodeLoop`: `taskA -> taskB`, setting `taskB -> taskA` returns `ErrCyclicDependency`.
  - `TestDetectCycles_DeepLoop`: 5-node chain `A -> B -> C -> D -> E`, setting `A.parent = E` returns `ErrCyclicDependency`.
  - `TestValidateHierarchyDepth_Subtree`: Moving a subtree where `parentDepth + 1 + subtreeDepth > 10` returns `ErrMaxDepthExceeded`.
  - `TestBuildTree_Forest`: Correctly groups multiple roots and nested child slices.
  - `TestBuildTree_Errors`: Orphaned `ParentID` returns `ErrTaskNotFound`; cyclic loops return `ErrCyclicDependency`.
- Verification: `go test -v -run "TestDetectCycles|TestBuildTree|TestValidateHierarchyDepth" ./internal/core/...`
- **Review Lenses**: Algorithmic correctness, performance.

### Unit 001-6: Task Filtering and Sorting Engine
- **Goal**: Implement in-memory filter evaluation and multi-key deterministic sorting.
- **Files**:
  - `internal/core/filter.go`
  - `internal/core/filter_test.go`
- **Dependencies**: Unit 001-4, Unit 001-5.
  - `TestFilterTasks`: Matches by Status, Priority, Tags, RootOnly, ParentID, and substring SearchTerm in title.
  - `TestSortTasks_MultiKey`: Sort by DueDate ASC (nil DueDate sorts last), then Priority DESC, with deterministic ID tie-breaking.
- **Verification**: `go test -v -run "TestFilter|TestSort" ./internal/core/...`
- **Review Lenses**: Simplicity, deterministic behavior.

---

## 4. Verification Matrix & Scenario Registry

| Requirement | Unit | Scenario IDs | Evidence Tier |
| :--- | :--- | :--- | :--- |
| **Error Taxonomy** | Unit 001-1 | CORE-ERR-N1, CORE-ERR-B1 | Focused unit |
| **Status State Machine** | Unit 001-2 | CORE-STS-N1, CORE-STS-B1, CORE-STS-F1 | Focused unit |
| **Priority Weighting** | Unit 001-2 | CORE-PRI-N1, CORE-PRI-B1 | Focused unit |
| **Tag Normalization** | Unit 001-2 | CORE-TAG-N1, CORE-TAG-B1 | Focused unit |
| **Task Lifecycle & Dates** | Unit 001-3 | CORE-TSK-N1, CORE-TSK-B1, CORE-TSK-R1 | Focused unit |
| **Progress Rollup Math** | Unit 001-4 | CORE-ROL-N1, CORE-ROL-B1, CORE-ROL-P1 | Focused property/table |
| **Cycle & Tree Invariants** | Unit 001-5 | CORE-TRE-N1, CORE-TRE-B1, CORE-TRE-F1, CORE-TRE-BM1 | Focused graph/benchmark |
| **Filtering & Sorting** | Unit 001-6 | CORE-FLT-N1, CORE-FLT-B1, CORE-FLT-C1 | Focused unit |
| **Aggregate Domain Suite** | All | CORE-AGG-ALL | Aggregate `make validate` (domain coverage threshold $\ge 95\%$ target; 98.0% measured) |

### Scenario Mapping Registry

| Scenario ID | Target Behavior | Executable Test Name | Command |
| :--- | :--- | :--- | :--- |
| `CORE-ERR-N1` | Unique sentinel errors with non-empty error strings | `TestErrorsExist` | `go test -v -run TestErrors ./internal/core/...` |
| `CORE-ERR-B1` | Identity comparisons work with `errors.Is` | `TestErrors_SentinelIntegrity` | `go test -v -run TestErrors ./internal/core/...` |
| `CORE-STS-N1` | Parse valid status enum strings | `TestParseStatus` | `go test -v -run TestParseStatus ./internal/core/...` |
| `CORE-STS-B1` | Invalid status string returns `ErrInvalidStatus` | `TestParseStatus_Invalid` | `go test -v -run TestParseStatus ./internal/core/...` |
| `CORE-STS-F1` | Reopening allowed, done -> blocked rejected | `TestStatusTransitions` | `go test -v -run TestStatusTransitions ./internal/core/...` |
| `CORE-PRI-N1` | Parse valid priority weights 1–4 and names | `TestParsePriority` | `go test -v -run TestParsePriority ./internal/core/...` |
| `CORE-PRI-B1` | Invalid priority integers/names return `ErrInvalidPriority` | `TestParsePriority_Invalid` | `go test -v -run TestParsePriority ./internal/core/...` |
| `CORE-TAG-N1` | Normalization lowercases, trims `#`, strips whitespace, dedupes | `TestNormalizeTag` | `go test -v -run TestNormalizeTag ./internal/core/...` |
| `CORE-TAG-B1` | Tags with invalid characters return `ErrInvalidTag` | `TestNormalizeTag_Invalid` | `go test -v -run TestNormalizeTag ./internal/core/...` |
| `CORE-TSK-N1` | NewTask validates title length 1–255, sets initial timestamps | `TestNewTask_Validation` | `go test -v -run TestNewTask ./internal/core/...` |
| `CORE-TSK-B1` | TransitionTo sets/clears CompletedAt appropriately | `TestTask_TransitionToDone_And_Reopen` | `go test -v -run TestTask ./internal/core/...` |
| `CORE-TSK-R1` | SetParent updates UpdatedAt; self-parenting returns ErrSelfParenting | `TestTask_SetParent` | `go test -v -run TestTask ./internal/core/...` |
| `CORE-ROL-N1` | Leaf task progress preserves manual progress or 100 on done | `TestCalculateProgress_Leaf` | `go test -v -run TestCalculateProgress ./internal/core/...` |
| `CORE-ROL-B1` | Floor integer arithmetic rounds down proportionally | `TestCalculateProgress_FloorRounding` | `go test -v -run TestCalculateProgress ./internal/core/...` |
| `CORE-ROL-P1` | Rollup average of subtasks clamped to 0–100 | `TestCalculateProgress_Subtasks`, `TestCalculateProgress_AllDone` | `go test -v -run TestCalculateProgress ./internal/core/...` |
| `CORE-TRE-N1` | BuildTree groups roots and children into hierarchical forest | `TestBuildTree_Forest` | `go test -v -run TestBuildTree ./internal/core/...` |
| `CORE-TRE-B1` | Root promotion with nil proposedParentID succeeds; subtree depth validated | `TestDetectCycles_RootPromotion`, `TestValidateHierarchyDepth_Subtree` | `go test -v -run "TestDetectCycles\|TestValidateHierarchyDepth" ./internal/core/...` |
| `CORE-TRE-F1` | Cyclic references return ErrCyclicDependency; orphans return ErrTaskNotFound | `TestDetectCycles_TwoNodeLoop`, `TestDetectCycles_DeepLoop`, `TestBuildTree_Errors` | `go test -v -run "TestDetectCycles\|TestBuildTree" ./internal/core/...` |
| `CORE-TRE-BM1` | Tree traversal micro-benchmark scales under 1000 nodes | `BenchmarkTreeTraversal` | `go test -bench=BenchmarkTreeTraversal ./internal/core/...` |
| `CORE-FLT-N1` | FilterTasks evaluates Status, Priority, Tags, SearchTerm, and RootOnly | `TestFilterTasks` | `go test -v -run TestFilterTasks ./internal/core/...` |
| `CORE-FLT-B1` | SortTasks sorts nil DueDate last on ASC, with deterministic ID tie-breaking | `TestSortTasks_MultiKey` | `go test -v -run TestSortTasks ./internal/core/...` |
| `CORE-FLT-C1` | Zero-match queries return empty non-nil slices | `TestFilterTasks_EmptyResults` | `go test -v -run TestFilterTasks ./internal/core/...` |
| `CORE-AGG-ALL` | Full test suite, race detector, static analysis, $\ge 95\%$ domain coverage target | All tests in `internal/core` | `make validate && go test -cover -race ./internal/core/...` |
