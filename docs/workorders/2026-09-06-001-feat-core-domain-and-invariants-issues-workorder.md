---
feature-id: 2026-09-06-001-feat-core-domain-and-invariants
plan-source: docs/plans/2026-09-06-001-feat-core-domain-and-invariants-plan.md
verification-plan: docs/verification-plans/2026-09-06-001-feat-core-domain-and-invariants-verification-plan.md
status: Open
evidence-scope: Planning findings only
---

# Feature 001: Core Domain & Invariants Issue Workorder

## 1. Issue Register

| ID | Source | Owner / Lens | Severity | Status | Impact | Next Action | Retest / Evidence |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **CORE-ISS-001** | Deepening Audit | Correctness / State Machine | P1 | Fixed in Plan | Legacy allowed invalid state jumps (`done -> blocked`). | Defined explicit `CanTransitionTo` matrix in plan. | Unit 001-2 test suite (`TestStatusTransitions`). |
| **CORE-ISS-002** | Deepening Audit | Performance / Graph Theory | P2 | Fixed in Plan | Legacy had no cycle detection on subtask reparenting, risking infinite UI loops. | Specified `DetectCycles` with depth limits in `internal/core/tree.go`. | Unit 001-5 test suite (`TestDetectCycles`). |
| **CORE-ISS-003** | Deepening Audit | Data Integrity / Rollup Math | P2 | Fixed in Plan | Ambiguity in progress integer rounding when division has fractional remainder. | Defined integer floor policy ($\lfloor \frac{\sum P}{N} \rfloor$). | Unit 001-4 test suite (`TestCalculateProgress`). |
| **CORE-ISS-004** | Deepening Audit | Domain Boundaries / Coupling | P1 | Fixed in Plan | Legacy mixed natural language date parsing inside domain entities. | Clarified that natural date parsing belongs to Feature 003 service layer; `Task` accepts pure `time.Time`. | Architecture boundary review in Unit 001-3. |

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

- [ ] Planned units implemented.
- [ ] Focused unit tests pass (`go test -v ./internal/core/...`).
- [ ] Race detector checks pass (`go test -race -v ./internal/core/...`).
- [ ] Benchmarks pass with expected performance (`go test -bench=. ./internal/core/...`).
- [ ] Aggregate repository validation passes (`make validate`).
- [ ] All issues fixed, or blocked issues carry owner, external blocker, revisit date, and user acceptance.
- [ ] Remaining unaccepted issues: 0.
