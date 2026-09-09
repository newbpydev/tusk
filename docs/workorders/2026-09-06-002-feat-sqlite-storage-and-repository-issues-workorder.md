---
feature-id: "002"
plan-source: docs/plans/2026-09-06-002-feat-sqlite-storage-and-repository-plan.md
verification-plan: docs/verification-plans/2026-09-06-002-feat-sqlite-storage-and-repository-verification-plan.md
status: locally accepted - native and hosted release gates deferred
evidence-scope: local implementation, per-unit commits and acceptance; no native or hosted execution
---

# Feature 002 Issue Workorder

This register accompanies the [implementation plan](../plans/2026-09-06-002-feat-sqlite-storage-and-repository-plan.md) and [verification plan](../verification-plans/2026-09-06-002-feat-sqlite-storage-and-repository-verification-plan.md). **Fixed in plan** means a planning gap was resolved in these documents, not that code was fixed or tests passed.

The original request authorized completing and reviewing the Feature 002 planning pack. The subsequent `ce-work` invocation authorizes implementation in masterplan order. Go 1.25 is the selected technical default; U6 owns runtime compatibility proof before migration implementation. Planning resolutions alone close no runtime gate; the execution receipts below record separately verified local acceptance.

## Issue register

| ID | Source / lens | Owner | Severity | Status | Next action and evidence |
| --- | --- | --- | --- | --- | --- |
| 002-ISS-022 | U5 recovery: migration failures lost context causes | Feature 002 implementer | P2 | Resolved locally | TestMigrate_PreservesCancellationCause red/green; safe causes survive DDL/ledger/commit/rollback and public mapping. |
| 002-ISS-023 | U5 tooling: plain make selected tool download | Feature 002 implementer | P2 | Resolved locally | Default-goal regression failed before .DEFAULT_GOAL := all; 12/12 base script checks now pass. |
| 002-ISS-024 | ce-code-review #1: schema category hid unknown cleanup | Feature 002 implementer | P1 | Resolved locally | TestOpenCause_RetainsMigrationUnknownOutcome failed before public mapper fix; uncertainty and safe cause now survive. |
| 002-ISS-025 | ce-code-review #2: inspection mislabeled cancellation | Feature 002 implementer | P2 | Resolved locally | TestInspection_PreservesCancellationCause failed for ledger/replay; cancellation and deadline categories now survive. |
| 002-ISS-021 | Runtime compatibility: lock wait ignores short context deadline | Feature 002 implementer; U6 | P1 | Resolved locally | KTD4 wrapper passes original cancellation regression, replacement/negative/recovery tests, Go 1.25 compatibility, five-target builds and make validate; see resolution receipt. |
| 002-ISS-001 | Coherence: Readiness advertised without an executable pack | Feature 002 implementer / named review lens | P1 | Fixed in plan | Set execution: code only with the complete contract, mapped scenarios and reviewed triplet; masterplan points to the first pending unit. Evidence: R21/R22; all units; 002-V67/002-V68. |
| 002-ISS-002 | Feasibility/dependencies: Go and SQLite minimum was unresolved | Feature 002 implementer / named review lens | P1 | Fixed in plan | Select Go 1.25.0, sqlite v1.58.0, libc v1.75.6 as the technical planning default; U6 must prove engine/graph/minimum compiler before U1. Evidence: R1; U6; 002-V01/002-V02/002-V07. |
| 002-ISS-003 | Feasibility: Generator minimum would contaminate runtime tooling | Feature 002 implementer / named review lens | P1 | Fixed in plan | Pin the prebuilt executable separately and commit generated output; ordinary build/test uses no source-built generator. Evidence: R18; U2; 002-V25/002-V28. |
| 002-ISS-004 | Architecture: Migration asset ownership and mutable state | Feature 002 implementer / named review lens | P2 | Fixed in plan | Place compiler-populated read-only embed.FS in db/embed.go; keep it private and never reassign it; test inventory rather than exempt db from coverage. Evidence: R7/R21; U1; 002-V09/002-V20. |
| 002-ISS-005 | Feasibility/portability: Checkout line endings could invalidate migrations | Feature 002 implementer / named review lens | P1 | Fixed in plan | U1 owns .gitattributes LF normalization for migration SQL; verify stable bytes and rejection of a deliberately changed digest. Evidence: R6/R7/R18; U1; 002-V12. |
| 002-ISS-006 | Security/migration: Foreign or newer databases could be changed during Open | Feature 002 implementer / named review lens | P1 | Fixed in plan | Require application ID, recognized schema and applied-prefix ledger; inspect read-only before writer/journal changes; distinguish transient SHM from main/WAL mutation. Evidence: R6/R7; U1/U3; 002-V13/002-V14/002-V39. |
| 002-ISS-007 | Data integrity: Migration failure had no atomic ledger contract | Feature 002 implementer / named review lens | P1 | Fixed in plan | One transaction owns identity, all pending SQL and ledger rows; preserve installed data and use inverse scripts only in disposable fixtures. Evidence: R6/R7/R16; U1; 002-V15/002-V16/002-V17. |
| 002-ISS-008 | Test strategy: Memory tests were incorrectly treated as WAL evidence | Feature 002 implementer / named review lens | P1 | Fixed in plan | Use unique memory names and an anchor for semantics; real temp disk and helper processes prove WAL/locking/recovery. Evidence: R19/R20; U3/U5; 002-V37/002-V60/002-V62/002-V63. |
| 002-ISS-009 | Security: Path overrides could be interpreted as DSNs | Feature 002 implementer / named review lens | P1 | Fixed in plan | Encode literal paths, reject unsafe final targets, inject lookup inputs and prove no fallback; preserve existing permissions. Evidence: R3/R4; U3; 002-V29/002-V30/002-V31/002-V32/002-V35. |
| 002-ISS-010 | Feasibility/concurrency: Reader flags and replacement pragmas were underspecified | Feature 002 implementer / named review lens | P1 | Fixed in plan | Use query_only readers and per-connection DSNs; writer IMMEDIATE mode and every replacement are compatibility probes. Evidence: R5/R14/R20; U6/U3; 002-V04/002-V05/002-V08/002-V36. |
| 002-ISS-011 | Adversarial/reliability: Commit failure was ambiguous about persisted state | Feature 002 implementer / named review lens | P1 | Fixed in plan | Return unknown outcome after uncertain commit/rollback, discard poisoned physical connections and require readback; never replay callbacks. Evidence: R14/R16; U4/U5; 002-V52/002-V55/002-V56/002-V57/002-V64. |
| 002-ISS-012 | Adversarial: Ignored callback operation errors could commit partial work | Feature 002 implementer / named review lens | P1 | Fixed in plan | Latch the first write-handle operation error; refuse later operations and roll back even if the callback returns nil. Evidence: R14/R15/R16; U4; 002-V54. |
| 002-ISS-013 | Correctness: Stored row decoding could apply domain defaults | Feature 002 implementer / named review lens | P1 | Fixed in plan | Use strict canonical codecs, validate whole results, preserve NULL/empty distinction and reject corruption without defaults or repair. Evidence: R8/R9/R11; U4; 002-V40/002-V41/002-V42/002-V43. |
| 002-ISS-014 | Coherence/correctness: SQL filters could diverge from core semantics | Feature 002 implementer / named review lens | P1 | Fixed in plan | Use a bound candidate superset then exact core filtering/sorting; invalid values do not match but valid alternatives still may. Evidence: R12/R13; U2/U4; 002-V23/002-V24/002-V44/002-V45. |
| 002-ISS-015 | Reliability/API: Transaction misuse lacked lifetime and failure rules | Feature 002 implementer / named review lens | P1 | Fixed in plan | Validate callbacks, carry transaction context, guard handles, reject nesting/concurrent use and wait for admitted operations before cleanup. Evidence: R15/R16; U4; 002-V53/002-V57. |
| 002-ISS-016 | Scope/data integrity: Storage acceptance could falsely claim business invariants | Feature 002 implementer / named review lens | P1 | Fixed in plan | Define the trusted storage boundary and structural delete safeguard; keep business validation and final service integration in Phase 3. Evidence: R13/R17/R22; U4/U5; 002-V19/002-V47/002-V49/002-V68. |
| 002-ISS-017 | Security/reproducibility: Tool installation and generation could leave partial artifacts | Feature 002 implementer / named review lens | P1 | Fixed in plan | Use release asset digests, safe single-executable extraction and scratch generation; compare entire output sets without Git dependence. Evidence: R18; U2; 002-V25/002-V26/002-V27/002-V28. |
| 002-ISS-018 | Reliability: Close could reject admitted work or wait for itself | Feature 002 implementer / named review lens | P1 | Fixed in plan | Use instance-owned admission tracking; allow admitted handles to finish, reject new work and prohibit owner Close from a callback. Evidence: R5/R15/R16; U3/U4; 002-V38/002-V53. |
| 002-ISS-019 | Test evidence/operations: Verification claims and canonical commands were conflated | Feature 002 implementer / named review lens | P1 | Fixed in plan | Label new Make target owners; leave all software scenarios unchecked; separate local, minimum-toolchain, cross-build, native and hosted records. Evidence: R1/R18/R21/R22; U6/U2/U5; 002-V07/002-V33/002-V66/002-V67/002-V68. |
| 002-ISS-020 | Architecture/sequencing: Migration unit depended on an unproved driver foundation | Feature 002 implementer / named review lens | P1 | Fixed in plan | Add stable unit U6/002-6 before U1; retain original U1–U5 IDs and synchronize product handoff, triplets and masterplan. Evidence: R1/R14/R21; U6/U1; 002-V01–002-V08. |

## Issue details

### 002-ISS-001. Readiness advertised without an executable pack

- **Found:** Planning/source review on 2026-09-08; P1; Coherence.
- **Owner:** Feature 002 implementer; Metadata and original outline.
- **Affected contract / retest:** R21/R22; all units; 002-V67/002-V68.
- **Evidence / planning gap:** The original frontmatter claimed implementation-ready while execution mode, concrete units and both companion artifacts were absent.
- **Correction:** Set execution: code only with the complete contract, mapped scenarios and reviewed triplet; masterplan points to the first pending unit.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-002. Go and SQLite minimum was unresolved

- **Found:** Planning/source review on 2026-09-08; P1; Feasibility/dependencies.
- **Owner:** Feature 002 implementer; Product G2 and KTD1.
- **Affected contract / retest:** R1; U6; 002-V01/002-V02/002-V07.
- **Evidence / planning gap:** The module declares Go 1.24, while the selected patched driver and libc manifests declare Go 1.25.
- **Correction:** Select Go 1.25.0, sqlite v1.58.0, libc v1.75.6 as the technical planning default; U6 must prove engine/graph/minimum compiler before U1.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-003. Generator minimum would contaminate runtime tooling

- **Found:** Planning/source review on 2026-09-08; P1; Feasibility.
- **Owner:** Feature 002 implementer; KTD2 and generator module.
- **Affected contract / retest:** R18; U2; 002-V25/002-V28.
- **Evidence / planning gap:** sqlc v1.31.1 source requires Go 1.26; putting it in the runtime module would undermine the Go 1.25 minimum.
- **Correction:** Pin the prebuilt executable separately and commit generated output; ordinary build/test uses no source-built generator.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-004. Migration asset ownership and mutable state

- **Found:** Planning/source review on 2026-09-08; P2; Architecture.
- **Owner:** Feature 002 implementer; KTD6 and db package.
- **Affected contract / retest:** R7/R21; U1; 002-V09/002-V20.
- **Evidence / planning gap:** Embedding from storage cannot reach ../../db and an exported asset variable would weaken state ownership.
- **Correction:** Place compiler-populated read-only embed.FS in db/embed.go; keep it private and never reassign it; test inventory rather than exempt db from coverage.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-005. Checkout line endings could invalidate migrations

- **Found:** Planning/source review on 2026-09-08; P1; Feasibility/portability.
- **Owner:** Feature 002 implementer; Schema and record codecs.
- **Affected contract / retest:** R6/R7/R18; U1; 002-V12.
- **Evidence / planning gap:** Applied SQL bytes never change, but the initial draft had no rule preventing Windows checkout from converting LF to CRLF.
- **Correction:** U1 owns .gitattributes LF normalization for migration SQL; verify stable bytes and rejection of a deliberately changed digest.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-006. Foreign or newer databases could be changed during Open

- **Found:** Planning/source review on 2026-09-08; P1; Security/migration.
- **Owner:** Feature 002 implementer; Open lifecycle and AE2.
- **Affected contract / retest:** R6/R7; U1/U3; 002-V13/002-V14/002-V39.
- **Evidence / planning gap:** The outline ran migrations at initialization without an application identity or read-only compatibility guard.
- **Correction:** Require application ID, recognized schema and applied-prefix ledger; inspect read-only before writer/journal changes; distinguish transient SHM from main/WAL mutation.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-007. Migration failure had no atomic ledger contract

- **Found:** Planning/source review on 2026-09-08; P1; Data integrity.
- **Owner:** Feature 002 implementer; KTD6 and migration lifecycle.
- **Affected contract / retest:** R6/R7/R16; U1; 002-V15/002-V16/002-V17.
- **Evidence / planning gap:** Clean rollback was asserted without a ledger schema, changed-checksum handling or failure at the ledger/commit boundary.
- **Correction:** One transaction owns identity, all pending SQL and ledger rows; preserve installed data and use inverse scripts only in disposable fixtures.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-008. Memory tests were incorrectly treated as WAL evidence

- **Found:** Planning/source review on 2026-09-08; P1; Test strategy.
- **Owner:** Feature 002 implementer; Original verification scenario 4.
- **Affected contract / retest:** R19/R20; U3/U5; 002-V37/002-V60/002-V62/002-V63.
- **Evidence / planning gap:** The same anonymous shared-memory URI appeared as the concurrency test environment despite memory journal differences and fixture collisions.
- **Correction:** Use unique memory names and an anchor for semantics; real temp disk and helper processes prove WAL/locking/recovery.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-009. Path overrides could be interpreted as DSNs

- **Found:** Planning/source review on 2026-09-08; P1; Security.
- **Owner:** Feature 002 implementer; R3/R4 and path lifecycle.
- **Affected contract / retest:** R3/R4; U3; 002-V29/002-V30/002-V31/002-V32/002-V35.
- **Evidence / planning gap:** The outline did not separate user filesystem paths from internal memory/connection URI parameters.
- **Correction:** Encode literal paths, reject unsafe final targets, inject lookup inputs and prove no fallback; preserve existing permissions.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-010. Reader flags and replacement pragmas were underspecified

- **Found:** Planning/source review on 2026-09-08; P1; Feasibility/concurrency.
- **Owner:** Feature 002 implementer; Driver tx.go and KTD4.
- **Affected contract / retest:** R5/R14/R20; U6/U3; 002-V04/002-V05/002-V08/002-V36.
- **Evidence / planning gap:** Driver ReadOnly only changes begin mode; new pooled connections do not inherit another connection's PRAGMA state.
- **Correction:** Use query_only readers and per-connection DSNs; writer IMMEDIATE mode and every replacement are compatibility probes.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-011. Commit failure was ambiguous about persisted state

- **Found:** Planning/source review on 2026-09-08; P1; Adversarial/reliability.
- **Owner:** Feature 002 implementer; Error and outcome contract.
- **Affected contract / retest:** R14/R16; U4/U5; 002-V52/002-V55/002-V56/002-V57/002-V64.
- **Evidence / planning gap:** A driver may commit before an injected acknowledgment failure; an error does not prove no mutation happened.
- **Correction:** Return unknown outcome after uncertain commit/rollback, discard poisoned physical connections and require readback; never replay callbacks.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-012. Ignored callback operation errors could commit partial work

- **Found:** Planning/source review on 2026-09-08; P1; Adversarial.
- **Owner:** Feature 002 implementer; Repository boundary.
- **Affected contract / retest:** R14/R15/R16; U4; 002-V54.
- **Evidence / planning gap:** The first draft rolled back callback errors but did not define an ignored failed event append followed by callback nil.
- **Correction:** Latch the first write-handle operation error; refuse later operations and roll back even if the callback returns nil.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-013. Stored row decoding could apply domain defaults

- **Found:** Planning/source review on 2026-09-08; P1; Correctness.
- **Owner:** Feature 002 implementer; KTD5 and current core/task.go.
- **Affected contract / retest:** R8/R9/R11; U4; 002-V40/002-V41/002-V42/002-V43.
- **Evidence / planning gap:** NewTask supplies defaults and may consult time.Now; it cannot safely decode persisted state without changing meaning.
- **Correction:** Use strict canonical codecs, validate whole results, preserve NULL/empty distinction and reject corruption without defaults or repair.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-014. SQL filters could diverge from core semantics

- **Found:** Planning/source review on 2026-09-08; P1; Coherence/correctness.
- **Owner:** Feature 002 implementer; List contract and core/filter.go.
- **Affected contract / retest:** R12/R13; U2/U4; 002-V23/002-V24/002-V44/002-V45.
- **Evidence / planning gap:** LIKE/NOCASE and blanket rejection of mixed invalid enum filters do not match the shipped core filter.
- **Correction:** Use a bound candidate superset then exact core filtering/sorting; invalid values do not match but valid alternatives still may.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-015. Transaction misuse lacked lifetime and failure rules

- **Found:** Planning/source review on 2026-09-08; P1; Reliability/API.
- **Owner:** Feature 002 implementer; Repository boundary and verification 002-V53.
- **Affected contract / retest:** R15/R16; U4; 002-V53/002-V57.
- **Evidence / planning gap:** Nil callbacks, retained handles, concurrent calls and callback completion during an admitted operation were not all specified.
- **Correction:** Validate callbacks, carry transaction context, guard handles, reject nesting/concurrent use and wait for admitted operations before cleanup.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-016. Storage acceptance could falsely claim business invariants

- **Found:** Planning/source review on 2026-09-08; P1; Scope/data integrity.
- **Owner:** Feature 002 implementer; R17 and product Phase 3 ownership.
- **Affected contract / retest:** R13/R17/R22; U4/U5; 002-V19/002-V47/002-V49/002-V68.
- **Evidence / planning gap:** FK/CHECK and traversal tests cannot prove service rollup, opposite-move prevention, stale edits or confirmation semantics.
- **Correction:** Define the trusted storage boundary and structural delete safeguard; keep business validation and final service integration in Phase 3.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-017. Tool installation and generation could leave partial artifacts

- **Found:** Planning/source review on 2026-09-08; P1; Security/reproducibility.
- **Owner:** Feature 002 implementer; Generator and build tooling.
- **Affected contract / retest:** R18; U2; 002-V25/002-V26/002-V27/002-V28.
- **Evidence / planning gap:** The outline had no executable pin/digest or non-mutating stale-output check, and Git diff would miss untracked generated files.
- **Correction:** Use release asset digests, safe single-executable extraction and scratch generation; compare entire output sets without Git dependence.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-018. Close could reject admitted work or wait for itself

- **Found:** Planning/source review on 2026-09-08; P1; Reliability.
- **Owner:** Feature 002 implementer; Connection and filesystem lifecycle.
- **Affected contract / retest:** R5/R15/R16; U3/U4; 002-V38/002-V53.
- **Evidence / planning gap:** Marking the repository closed before admitted callbacks finish needs an admission distinction, and Close inside a callback deadlocks.
- **Correction:** Use instance-owned admission tracking; allow admitted handles to finish, reject new work and prohibit owner Close from a callback.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-019. Verification claims and canonical commands were conflated

- **Found:** Planning/source review on 2026-09-08; P1; Test evidence/operations.
- **Owner:** Feature 002 implementer; Verification Contract and baseline Makefile.
- **Affected contract / retest:** R1/R18/R21/R22; U6/U2/U5; 002-V07/002-V33/002-V66/002-V67/002-V68.
- **Evidence / planning gap:** Future targets were absent, core history was not fresh evidence, and cross-builds cannot establish native platform or CLI latency acceptance.
- **Correction:** Label new Make target owners; leave all software scenarios unchecked; separate local, minimum-toolchain, cross-build, native and hosted records.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

### 002-ISS-020. Migration unit depended on an unproved driver foundation

- **Found:** Planning/source review on 2026-09-08; P1; Architecture/sequencing.
- **Owner:** Feature 002 implementer; Original U1 compatibility work and phase order.
- **Affected contract / retest:** R1/R14/R21; U6/U1; 002-V01–002-V08.
- **Evidence / planning gap:** The original five-unit split folded driver selection/proof into schema work with no independently verifiable prerequisite.
- **Correction:** Add stable unit U6/002-6 before U1; retain original U1–U5 IDs and synchronize product handoff, triplets and masterplan.
- **Status:** Fixed in plan. Review the owning plan section and cited scenario for closure; runtime execution remains pending.

## Review coverage and planning receipt

The composing workflow is ce-plan → plan-ultrathink deepening → ce-doc-review in non-interactive mode → synchronized pack audit. The full document-review workflow was adapted to the repository's explicit instruction to run subagent work sequentially in the main thread. No subagents or external model processes were dispatched, and no independence/consensus promotion is claimed.

| Pass / lens | Scope and reason | Result | Follow-up evidence |
| --- | --- | --- | --- |
| ce-plan repository grounding | Live core, Makefile, active master target, clean HEAD 21bb210; no storage implementation exists | Captured in Planning Contract | Recheck facts before implementation |
| ce-plan external grounding | Driver/libc manifests, SQLite behavior, sqlc source minimum and release digests | Pins and failure semantics decided | U6 compatibility and U2 generator tests |
| Deepening: KTDs and dependencies | Initial outline had driver/tool alternatives and missing prerequisites | U6 split, KTD1–KTD10 selected | 002-ISS-002/003/020 |
| Deepening: implementation/verification | Units lacked ownership, red-first proof and concrete failure cases | Six units and 68 scenarios mapped | All scenarios remain not executed |
| Deepening: risks and operations | Foreign DB, schema drift, WAL/cancellation and native evidence missing | Lifecycle, error and handoff contracts added | U1/U3/U5 gates |
| ce-doc-review coherence | Always-on; full unified-plan and both companions | Corrected filter wording and callback/verification agreement | 002-ISS-014/015 |
| ce-doc-review feasibility | Always-on; compare design with driver/source/tool manifests | Corrected migration byte portability | 002-ISS-005 |
| ce-doc-review scope guardian | 22 requirements and future-phase boundaries | Business rules and host acceptance stay with owners | 002-ISS-016/019 |
| ce-doc-review security | File-access boundary and downloaded executable trust | Literal paths, restrictive new files, digest checks and error redaction covered | 002-ISS-006/009/017 |
| ce-doc-review adversarial | Data migrations and new transaction abstraction | Corrected swallowed-operation failure and Close admission | 002-ISS-012/018 |
| Data integrity / migration | Required ultrathink persistence lens | Atomic ledger, events and rollback/readback covered | U1/U4/U5 scenarios |
| Test strategy / performance | Required evidence and latency ownership lens | Real disk evidence; local benchmarks do not claim CLI/native success | 002-ISS-008/019 |
| Simplicity / maintainability | Required core lens | No ORM, custom worker pool, network analyzer, cache or migration framework | Reassess abstractions during implementation |
| Cross-model corroboration | Would activate for adversarial/security | Not run: project requires sequential main-thread review | No independent-review claim |
| Product/design personas | No new product-position decision or UI behavior in this feature | Not activated; product traceability checked above | Later service/CLI/TUI plans own their reviews |

The first review pass applied five concrete corrections to the drafted pack: LF-controlled migration hashes, latched transaction failures, Close admission rules, mixed enum filter semantics, and complete callback-lifetime rules. The broader register also retains the earlier outline/grounding findings. These are document fixes, not implementation results.

Non-interactive review envelope: five fixes applied; no remaining proposed fix or user-judgment finding against the selected technical plan; no cross-model result. Routine execution-time compatibility, performance and native-platform evidence are planned gates, not unresolved architecture questions. Final correction readback and structural audit are recorded below.

## Planning audit record

| Date | Scope | Method | Result |
| --- | --- | --- | --- |
| 2026-09-08 | Live baseline and official dependency/tool sources | Read-only inspection; no tests/install/build/generation | Captured in plan; runtime proof pending |
| 2026-09-08 | Final triplet and connected authority documents | Markdown/link/traceability/status audit through a temporary `planning-audit-002` recipe supplied to the canonical Makefile with `--eval`; no repository target added | Passed: nine Markdown/MDC files, 55 local links, 22 requirements, six ordered units, 68 unique unchecked scenarios and 20 issue records; product retains 29 requirements/73 unchecked scenarios and now has 24 units; documentation-only diff and `git diff --check` passed |
| 2026-09-08 | Correction readback and handoff | Rechecked the five draft corrections across contracts, scenarios and issue records; compared masterplan checks to HEAD | Planning 2.1 and its three artifact checks complete; next target 002-6 / U6; no implementation, scenario or release check advanced |

Planning checklist 2.1 is complete after the documentation audit. All six implementation units, every 002-V scenario, phase acceptance and product runtime gate G2 remain unexecuted. `make validate` was not run: this pass changed planning documents and storage guidance only, and makes no application verification claim.

## Implementation and release gate

### 002-ISS-021. Pinned driver's lock wait outlives the context deadline

- **Observed:** 2026-09-08 on clean base `4bbc639`, with uncommitted U6 changes on `feat/sqlite-storage-repository`.
- **Contract:** R16/R20, KTD4/KTD9; U6; 002-V05/002-V06. The plan explicitly makes failed compatibility proof a stop before U1.
- **Reproduction:** Two independently opened single-connection pools on a temporary disk WAL database, both configured with `_txlock=immediate` and `busy_timeout=5000`. Writer A holds a transaction; writer B calls `BeginTx` with a 100-ms deadline. B returns `database is locked (5) (SQLITE_BUSY)` after 5.03763051s on Go 1.25.0 and 5.009708107s on installed Go 1.27.1. Releasing A permits a later B transaction. The race run reproduces the same failure after 5.05168653s.
- **Source inspection:** Pinned [tx.go](https://github.com/modernc-org/sqlite/blob/v1.58.0/tx.go) executes begin through `sqlite3_exec`, arranges interruption for canceled contexts, and returns the SQLite error. The measured behavior does not satisfy the promised short-deadline lock cancellation; source inspection alone does not establish a safe fix.
- **Changes retained:** Pinned go.mod/go.sum, private no-I/O connector, compatibility regressions, stricter Go-minimum setup fixtures, and canonical compatibility/cross-build targets. No schema, public opener, repository, service, CLI or TUI implementation was added.
- **Validation:** Engine/minimum compiler and individual basic connection probes pass. `make test-compat`, `make validate` and separately `make race` fail the retained regression; coverage was not reached. The setup script suite passes 9/9. Full evidence is in the paired verification record.
- **Cross-build proof:** `GOTOOLCHAIN=go1.25.0 make build-storage` exits 0 for the native package and all five CGO-disabled targets. This does not close the failed runtime gate or establish native Windows/macOS behavior.
- **Resolution in progress:** Following the user's instruction to continue with engineering judgment, KTD4 now selects a connection-local cancellable acquisition wrapper. The default busy timeout and total lock budget remain five seconds; only acquisition temporarily uses 25-ms waits, with context checks between failed driver BeginTx attempts. Restore the default before returning a connection/transaction; discard on restoration failure. No manual SQL BEGIN, dependency change, callback/statement replay or relaxed cancellation assertion. Added negative/recovery tests must prove this before closing the issue.
- **Resume:** Resolve this incompatibility, finish the remaining U6 mode/portability probes, and pass minimum/current `make test-compat`, `make build-storage`, full/race and `make validate`. Until then, U6 and all downstream units remain unchecked; no commit or publication performed.

### Gate checklist

- [x] U6, U1, U2, U3, U4 and U5 implemented in masterplan order.
- [x] Each behavioral change has observed red-first evidence and focused green proof.
- [x] Minimum Go/compiler and pinned SQLite/libc engine proof passed.
- [x] Generated output and negative script checks passed.
- [x] Full, race and required per-package coverage passed through Make.
- [x] Applicable local disk/process scenarios executed and recovery demonstrated.
- [x] Native Windows/macOS and hosted evidence recorded separately, or retained as named Phase 6 gates.
- [x] All runtime findings resolved with fresh retest evidence.
- [x] Phase 3 handoff and masterplan pointers synchronized.
- [x] Remaining unaccepted implementation issues: 0.

### 002-ISS-021 resolution receipt

Resolved locally on 2026-09-08. The connection wrapper retries only failed acquisition before user work, restores busy_timeout=5000, preserves all driver pool interfaces, and poisons a connection when restoration fails. Both cancellation regressions pass around 100 ms, and fault/budget/replacement tests pass. Explicit Go 1.25 compatibility and five-target builds pass; make validate passes with 97.8% storage coverage. Earlier failure/stop notes above remain historical evidence. U6 is accepted and U1 becomes active. No commit or publication occurred.

### U1 execution and review receipt

U1 accepted after final make validate on 2026-09-08 (storage 96.1%, db 100%). Data-integrity/migration review exercised atomic DDL+ledger+task rollback, foreign/newer/drift refusal and constraints. Adversarial tests reproduced and fixed two implementation findings: wildcard catalog exclusion (sqliteXsecret) and split-snapshot identity/catalog inspection. Failure injection distinguishes commit acknowledgment loss from pre-commit failure and proves recovery; two-process tests prove serialized initialization. Source inventory and disposable inverse fixtures are isolated. ce-simplify-code was performed inline under project instructions, with no edits needed. See paired verification receipt for exact test names. Masterplan advances to U2; final feature review and downstream acceptance remain pending.

### U2 execution and review receipt

U2 accepted on 2026-09-08 after make validate check-generated. Official prebuilt sqlc v1.31.1 installation verified the pinned digest; generated output is reproducible. SQL correctness/injection, recursive termination, candidate parity and tooling recovery fixtures pass. Resolved two observed implementation findings: ambiguous recursive ID references and coverage-script package-column parsing. Existing generated-code exemptions were preserved, not broadened. Whole-directory checks work without Git; generation and installation failures preserve old output/tool. Paired verification record owns details; masterplan advances to U3.

### U3 execution and review receipt

U3 locally accepted after make build-storage validate check-generated; storage coverage 95.7%. Filesystem/privacy, resource ownership, concurrency and portability checks are represented by the paired path/open/fault/memory tests. No implicit CLI database access was introduced. Every counted physical handle closes on injected failure and a subsequent Open succeeds. Native Windows 002-V33 remains a Phase 6 obligation. Masterplan advances to U4.

### U4 execution receipt — 2026-09-08

Base 4bbc639 plus uncommitted implementation. make validate check-generated passed; handwritten storage coverage 97.0%. Go 1.25 test-compat and five CGO-disabled storage/test builds passed. Repository round-trip, exact core filtering, tree corruption, deletion, history, snapshot, lifetime/concurrent-handle, cancellation, commit/rollback uncertainty and driver-fault tests cover 002-V40–002-V58. Observed regressions before fixes: public Open leaked OS paths; ListChildren unnecessarily decoded corrupt grandchildren; rollback cleanup failures lost the original context/domain cause. Sanitized categories, immediate-child reads and joined safe causes resolve those failures. Callback replay remains prohibited, provisional failed reads return no data, and the read handle exposes no writer interface. U4 is locally accepted; U5 is active. Native/hosted execution is not claimed.

### U4 commit reconstruction — 2026-09-08

The original red-first work was accumulated without per-unit commits. At the user's correction, this unit was reconstructed in an isolated worktree and make validate was rerun on its exact code contents before committing. The original chronological test receipts above remain historical evidence. Unit completion now includes a separate local commit before advancing; pushing and merging are outside this authorization.

### U5 acceptance receipt — 2026-09-08

All applicable Feature 002 local scenarios are accepted: 67/68 checked, with only native Windows 002-V33 deferred to Phase 6. Current Go 1.27.1-X:nodwarf5 make validate check-generated build passed after the review fixes (storage coverage 97.6%, db/cmd 100%, core 98.2%; existing ports/sqlc exemptions unchanged). Explicit Go 1.25 previously passed the full gate; after the final mapper changes it again passed test-compat, all five CGO-disabled storage/test builds, check-generated and focused migration/inspection race tests. make setup and the final Btrfs make bench-storage run passed. No hosted or native macOS/Windows runtime result is claimed.

Two helper processes preserve all 40 read-modify-write increments and 40 events; a reader retains its snapshot across another process's commit. External writer cancellation returns in 102.5 ms (105.3 ms under race), while the uncanceled wait returns busy in 5.01 s; later writes succeed. Killing and reaping a writer before commit preserves the old task/event set; killing after acknowledgment preserves exactly one committed set. Independent reopen checks integrity and foreign keys. Held-reader WAL growth is followed by automatic restart sequence 0 -> 5 with bounded file reuse, without explicit checkpoint SQL. Normally closed offline backup reopens without modifying its source.

Observed red-first U5 fixes restore the default Make target, preserve migration statement/ledger/commit/rollback cancellation causes, preserve inspection cancellation, and retain unknown migration outcomes alongside schema categories. The completed ce-code-review receipt reported two actionable findings; both were reproduced and fixed, with no unapplied actionable residual. Review passes ran sequentially in the parent context per repository tool mapping; both independent peer routes failed before producing a review, so independent corroboration is unavailable. ce-simplify-code found no warranted behavior-preserving edit.

[Durable evidence, commit sequence and review resolution](../verification-evidence/002/README.md), [raw benchmarks](../verification-evidence/002/storage-benchmarks.txt), [code fingerprint and gate receipt](../verification-evidence/002/acceptance.json), and [operations/service handoff](../storage.md) retain the evidence. The separate U5 commit closes Phase 2; the next active target is Phase 3 planning, not Feature 003 implementation. No push, PR, merge or release is authorized by this acceptance.

### Post-acceptance publication follow-up — 2026-09-08

The user authorized simplify, review to zero actionable findings, compound, then commit/push/PR. MASTERPLAN target 2.4 tracks this follow-up; the six validated implementation commits remain separate. Simplification found no worthwhile behavior-preserving changes. Fresh review reports zero actionable findings; make validate check-generated build passes with 97.6% storage coverage. Both external review routes failed before producing usable review evidence; nine local passes ran sequentially in the parent context. See ../verification-evidence/002/publication-review.json. The reusable transaction-outcome/redaction lesson is captured in ../solutions/database-issues/preserve-transaction-outcomes-through-error-redaction.md with parser, link and source grounding checks. The branch is published as [PR #2](https://github.com/newbpydev/tusk/pull/2) against main; merge remains user-owned. Hosted checks and feedback are handled by the PR monitor, separately from local acceptance. Feature 003 implementation remains out of scope; V33 native Windows remains a Phase 6 obligation.
