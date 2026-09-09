---
feature-id: "003"
plan-source: docs/plans/2026-09-06-003-feat-task-service-engine-plan.md
verification-plan: docs/verification-plans/2026-09-06-003-feat-task-service-engine-verification-plan.md
status: Planning complete - execution gates pending
evidence-scope: Planning findings only
---

# Feature 003 Task Service Engine Issue Workorder

Companions: [plan](../plans/2026-09-06-003-feat-task-service-engine-plan.md), [verification plan](../verification-plans/2026-09-06-003-feat-task-service-engine-verification-plan.md) and [masterplan](../../MASTERPLAN.md).

**Fixed in plan** closes a missing decision or planning contradiction only. It never claims that implementation or its tests passed. **Open execution gate** has a decided approach and requires the named future evidence; it does not block the prerequisite unit that creates that evidence. Owners below are project roles, not claims that named people accepted assignments. There are no unresolved product decisions blocking plan readiness. Application tests and make validate were not run in this planning pass.

## Issue register

Lfg intake, 2026-09-09: local sequential contract review found no additional blocking findings at `f2fd0e35f16edb78ef5e57c0f4a68ffa7a1b9fa0`. The optional Claude review returned HTTP 401 (expired OAuth token), so no independent review is claimed. Implementation and publication through an open PR are now user-authorized; execution gates retain their evidence requirements.

| ID | Source / owner and lens | Severity | Status | Impact / next action | Retest or closure evidence |
| --- | --- | --- | --- | --- | --- |
| 003-ISS-001 | Outline / plan owner, architecture | P1 | Fixed in plan | Unsupported readiness/missing companions; use synchronized seven-unit pack | Pack audit and master/registry/product G1 handoff |
| 003-ISS-002 | Product R18/R19 / service owner, API parity | P1 | Fixed in plan | Missing get/stats/history/forest/delete-preview contracts; adopt operation/type tables | U1/U4/U7; V03, V60, V65–V76 |
| 003-ISS-003 | Patch sketch / service owner, correctness | P1 | Fixed in plan | Preserve omit/set/clear, current-row patch and no-op semantics | U1/U6/U7; V04–V10, V44–V48 |
| 003-ISS-004 | Transaction boundary / storage-service owner, reliability | P1 | U1 verified; U5 aggregate pending | All new service causes survive failed rollback; safe categories added in U1 | V88; every new cause individually/joined, private text redacted and unknown outcome retained |
| 003-ISS-005 | Date outline / parser owner, portability | P1 | Fixed in plan | Decide local date/end time, calendar clamps, DST ambiguity, ranges and strict grammar | U2; V13–V24, V89 |
| 003-ISS-006 | Identity omission / service owner, API/security | P2 | Fixed in plan | Pin UUIDv7 input/entropy/layout/ranges, invalid injected IDs and no retry | U2/U6; V25–V27, V41 |
| 003-ISS-007 | Rollup status ambiguity / service owner, domain | P1 | Fixed in plan | Distinguish progress100 from done, explicit reopen from policy completion, and final-child reset | U3/U7; V28–V35, V49–V59 |
| 003-ISS-008 | Move traversal / service owner, integrity | P1 | Fixed in plan | Validate subtree height and compute shared chains only after final membership | U3/U7/U5; V34, V38, V56–V59, V78 |
| 003-ISS-009 | TUI base handoff / service owner, conflict | P1 | Fixed in plan | Explicit value equality, incarnation, parent-derived progress and ABA semantics | U1/U6/U5; V10, V46, V47, V79 |
| 003-ISS-010 | Delete sketch / service owner, destructive actions | P1 | Fixed in plan | Require consent snapshot or force, preserve recursion guard, recheck membership/metadata | U7/U5; V60–V64, V80 |
| 003-ISS-011 | Event ownership / service owner, privacy | P1 | Fixed in plan | Define net-diff attribution, ordering, create fields and deletion retention | U3/U6/U7/U5; V36, V37, V39, V62, V75, V87 |
| 003-ISS-012 | Query boundaries / service owner, correctness | P1 | Fixed in plan | Core-exclusive bounds cannot express inclusive midnight; use separate day predicate and complete ancestry | U4; V65–V76 |
| 003-ISS-013 | Mock-only acceptance / test owner, evidence | P1 | Open execution gate | Add real WAL, stale-write, per-statement and process termination service tests in U5 | V77–V87; disk snapshots, callback counts, integrity and raw logs |
| 003-ISS-014 | Broad implementation units / plan owner, sequencing | P2 | Fixed in plan | Split original orchestrator into U6 creation/patch and U7 lifecycle/deletion while preserving U1–U5 identities | Dependency/ownership audit; separate red/validate/commit boundaries during execution |
| 003-ISS-015 | Runtime/measurement / U5 owner, portability/performance | P2 | Open execution gate | Add tested canonical build-service/bench-service targets, minimum Go proof and workload baselines | V90/V91 plus make validate; no native/CLI latency claim |
| 003-ISS-016 | Consumer packaging / Feature 004–006 owners | P2 | Open handoff gate | Embed production timezone data, implement adapter contracts and prove native/hosted/terminal behavior | Owning phase triplets and exact candidate evidence; not local Feature 003 acceptance |

## Issue details

### 003-ISS-001. Readiness and triplet authority

- **Found / affected:** Planning, 2026-09-09; R22, all units. The 46-line outline marked itself implementation-ready without command types or companion artifacts.
- **Correction:** New R1–R22, KTD1–KTD10, seven ordered units and 91 planned scenarios now define the contract. MASTERPLAN alone tracks activation; product G1 closes for Feature 003 planning only.
- **Owner / closure:** Plan owner; document audit must validate IDs/links/coverage and preserve implementation/release checkboxes. No runtime inference from metadata.

### 003-ISS-002. Missing consumer operations

- **Found / affected:** Product/contract review; R2/R15/R17–R19, U1/U4/U7. Original methods omitted get, history, stats, forest and delete preview despite product requirements.
- **Correction:** Exact operation/type tables include detached results, non-nil collections, optional Base/Expected and structured deletion. Generic multi-ID batching remains outside product grammar.
- **Owner / retest:** Service API owner; compile consumer fixtures and run V03/V60/V65–V76. CLI wire DTOs remain Feature 004; no premature UI implementation.

### 003-ISS-003. Patch intent and no-op

- **Found / affected:** Correctness review; R3/R8/R14, U1/U6/U7. Whole-row updates can lose omitted fields; core setters may touch UpdatedAt even when values do not change.
- **Correction:** Copy supplied commands; pointer/clear rules; apply to authoritative row; canonical net diff ignores timestamp-only writes. Empty service patch is no-op, with Base validation; empty CLI edit remains later syntax error.
- **Owner / retest:** Service owner; set/omit/clear, aliasing, same-value and conflicting-directive cases V04–V10/V44–V48.

### 003-ISS-004. Failed rollback can lose service error causes

- **Found / affected:** Live source `internal/storage/transaction.go` and `errors.go`; R6, U1/U5, V88. Normal callback rollback returns the original error, but failed cleanup sanitizes against a finite list which does not yet include new service categories.
- **Root cause / correction:** Extend existing safe allowlist alongside new immutable port errors. Do not change the repository callback API or expose driver unwrap chains. A cause and unknown outcome are independent facts.
- **Owner / next action:** U1 storage/service boundary owner writes failing cases for every new sentinel, then minimal mapper extension. U5 repeats aggregate acceptance. Include joined causes and private wrapper text using existing fault patterns.
- **Blocking effect / evidence:** Blocks U1 completion and later service implementation until errors.Is/As tests and make validate pass; does not block starting U1. No runtime fix claimed. Revisit at U1 red-first execution; no waiver accepted.

### 003-ISS-005. Calendar behavior beyond token names

- **Found / affected:** Product R15/R16, Go time documentation; R5/R17, U2/U4. Elapsed-hour arithmetic fails DST, AddDate does not clamp months, and Go's ambiguous Date choice is unspecified.
- **Correction:** Fixed reference/location, strict grammar, direct destination-month clamping, separate [start,end) day query, range checks and explicit missing/repeated wall-time rules. Conservative rejection of missing historical midnight boundaries is a documented feature planning default, not claimed user approval.
- **Owner / retest:** Parser owner; V13–V24/V68/V89 cover normal, malformed, overflow, 23/25-hour day, skipped date and repeated wall time. No parser execution evidence exists yet.

### 003-ISS-006. Identity and failure injection

- **Found / affected:** Product UUIDv7 requirement versus outline omission; R4, U2/U6. No source of identity/time or collision response was specified.
- **Correction:** Inject timestamp/ID source; standard-library UUID helper with supplied entropy, exact version/variant and timestamp limits; canonical injected-ID validation; one attempt only.
- **Owner / retest:** Service owner; RFC vector, short/error entropy, negative/overflow clock, same millisecond and duplicate insert V25–V27/V41. ID time leakage is inherent to the chosen product UUID format; no authorization/security token use.

### 003-ISS-007. Status, progress and auto-completion

- **Found / affected:** Outline rollup and live core done-first behavior; R10–R14, U3/U7. Updating progress alone can leave done ancestors over open descendants or instantly undo an explicit reopen.
- **Correction:** Reopen done ancestors before calculating; status-open parent at 100 is still open. Suppress auto-completion for explicitly opened target. Require all direct statuses done, never just 100. Reset last-child open parent to 0.
- **Owner / retest:** Service/domain owner; V28–V35/V49–V59. Review final graph and no-op event behavior with both policies, including status-only changes at progress100.

### 003-ISS-008. Moves and shared-ancestor ordering

- **Found / affected:** Graph/data-integrity review; R9/R10, U3/U7/U5. Updating each chain separately can double-write a shared ancestor or observe a half-applied move.
- **Correction:** Validate target subtree height and destination ancestry under writer lock; stage membership; process final-depth union once. Cache staged children so reads are not taken from stale persisted membership before flush.
- **Owner / retest:** Service owner; V34/V38/V56–V59/V78 with shared/disjoint chains, descendant-related parent chains, depth10/11 and concurrent opposite moves. Rejected mutations preserve events too.

### 003-ISS-009. Stale form equality

- **Found / affected:** Product KTD12; R8, U1/U6/U5. Unspecified equality could ignore changed notes, compare pointers, or conflict on parent progress which is not editable.
- **Correction:** Compare complete editable values and CreatedAt incarnation; compare Progress only for current leaf. Ignore UpdatedAt/CompletedAt. Equal-value ABA is accepted; schema revisions are not introduced.
- **Owner / retest:** Service owner; V10/V46/V47/V79. Freshness is independent of no-op detection. Form draft retention/message rendering remains Feature 005.

### 003-ISS-010. Consent-to-delete race

- **Found / affected:** Product R20 versus bare recursive bool signature; R15, U7/U5. A displayed count does not identify confirmed membership, and force could accidentally imply recursion.
- **Correction:** Preview target+sorted IDs in one snapshot; compare authoritative membership and metadata inside write; require Expected for nonforce. Force bypasses only consent and cannot be combined with Expected.
- **Owner / retest:** Service/deletion owner; V60–V64/V80. Membership or compared metadata changes require renewed confirmation; unrelated changes pass. No prompt lives in service.

### 003-ISS-011. History attribution and privacy

- **Found / affected:** Product R18 and current TaskEvent allowlist; R16, U3/U6/U7. Multiple internal setter transitions could emit duplicates or retain user content.
- **Correction:** Diff original/final values, assign each changed field once to a deterministic category, fixed task/category ordering and one timestamp. Creation lists initialization fields; deletion cascades without tombstones.
- **Owner / retest:** Service/privacy owner; V36/V37/V39/V62/V75/V87. Assertions include exact fields, timestamp/sequence behavior, no-op, no raw old/new text and unrelated-history preservation.

### 003-ISS-012. Query, tree and statistic contracts

- **Found / affected:** Core filter/tree APIs; R17–R19, U4. Exclusive due bounds lose exact midnight if reused directly; filtered/selected subtree without ancestors is not a valid BuildTree input.
- **Correction:** Separate half-open day predicate, explicit default status logic, full ancestry before detached depth projection, retained-task statistic formulas and all-events history.
- **Owner / retest:** Query owner; V65–V76. No custom sort/pagination framework or fabricated timeline. Errors from read cleanup suppress all constructed results.

### 003-ISS-013. Real service transaction proof

- **Found / affected:** Original mock-only suite and storage acceptance boundary; R20, U5. Storage tests supplied rollup values; they do not prove service computation under concurrent writes.
- **Correction:** Real disk/two-owner service tests, per-statement/event writer decorators, process barriers and old/new complete graph/history readback. Simulated lost acknowledgment is labeled separately from storage driver fault coverage.
- **Owner / next action:** U5 test owner implements V77–V87 after service exists. Capture race/full/coverage receipts and readback; reuse public storage.Open only.
- **Blocking effect / evidence:** Blocks Phase 3 local acceptance, not earlier units. Revisit at U5; no waiver accepted. Child processes always reaped; no user DB or sidecar deletion.

### 003-ISS-014. Unit granularity and commit boundaries

- **Found / affected:** Master outline 003-4 and late test-only unit; R22, all units. A single all-method orchestrator hides validation and concurrency dependencies.
- **Correction:** Preserve U1–U5 IDs, add U6/U7 for creation/patch and lifecycle/deletion, make U4 final queries/facade and U5 integration. Tests start within each behavior unit; U5 is not the first time service tests appear.
- **Owner / retest:** Plan owner verifies dependency DAG and ownership; future executor supplies separate red/green/validate and commit per unit. No blanket staging or commit occurs in planning.

### 003-ISS-015. Compatibility and performance evidence

- **Found / affected:** Makefile and coverage script; R21/R22, U5. There is no build-service/bench-service target or focused test variable support today.
- **Correction:** Use existing containing Make suites; U5 adds/tests explicit service targets and records minimum-Go/CGO0 cross-builds, 0/100/1000/10000-task measurements and runbook replay.
- **Owner / next action:** U5 implementer; V90/V91 plus full make validate and raw host/revision/fixture output.
- **Blocking effect:** Blocks Phase 3 acceptance until evidence exists; CLI timing and native runtime remain separate gates. Revisit U5; no latency waiver or current performance claim.

### 003-ISS-016. Consumer and release handoffs

- **Found / affected:** Product F3/R21–R28/G3/G4; R21, U2/U5. Library date tests cannot prove single-binary timezone data or real terminal/JSON behavior.
- **Correction:** Feature 004 owns production main-package time/tzdata embedding, API-to-CLI syntax/domain mapping, policy flag/env parsing, wire DTOs, consent and startup latency. Feature 005 owns forms/drafts/refresh and terminal acceptance. Feature 006 owns native/hosted builds/runtime/release.
- **Owner / next action:** Owning feature planner carries these into its triplet before implementation (G1). Revisit at Feature 004 planning and each later feature handoff.
- **Blocking effect / evidence:** Does not block local service acceptance, which has no runnable consumer. Blocks respective consumer/release acceptance. No explicit user acceptance of a release waiver is recorded or required to document a later-phase obligation.

## Review-lens sign-offs

Reviews run sequentially in the main thread under the repository's tool mapping. These are document-review results, not independent reviewer votes or implementation sign-offs. Each correction was checked against the plan and verification scenarios together.

| Lens | Planning result | Findings and corrections | Execution retest |
| --- | --- | --- | --- |
| Architecture/dependency sequencing | Reviewed | ISS-001/002/014: port ownership, seven-unit DAG, no incomplete public method stubs | U1–U7 in declared order |
| Product/scope and automation parity | Reviewed | ISS-002/007/010/016: full durable-action port, subtree bulk scope, no hidden prompt/config | API and consumer handoff scenarios |
| Correctness/reliability | Reviewed | ISS-003/004/005/007/012: no-op/time/rollup/error outcome rules | V04–V38, V44–V59, V65–V88 |
| API/library compatibility | Reviewed | ISS-002/003/009/010: command/result shapes, consent/base/clear defaults, error taxonomy | V03–V12, V44–V48, V60–V76 |
| Data integrity/concurrency | Reviewed | ISS-007/008/013: staged final graph, one transaction, deepest-first union, real disk proof | V28–V64, V77–V86 |
| Security/privacy | Reviewed | ISS-004/006/011: safe errors, no retry, metadata-only events, test isolation | V26/V37/V41/V83/V87/V88 |
| Portability/performance/operations | Reviewed with later gates | ISS-005/015/016: deterministic zone handling, minimum runtime and measured scope | V89–V91 plus Features 004–006 |
| Test strategy/evidence quality | Reviewed with execution gates | ISS-001/013/015: all requirements mapped, doubles distinguished from disk/driver faults, no false pass | All scenario receipts and canonical gates |
| Simplicity/maintainability | Reviewed | No global cache, custom workers, NLP/UUID dependency, new schema, general command bus or unused sorting API | Ownership/import/resource review at implementation |
| Documentation/coherence | Audited | Eight Markdown files, 97 local links/anchors, 22 requirements, seven units, 91 scenarios, 16 issues, status and master/product synchrony | Rerun structural audit after contract changes |

## Planning evidence record

- **Inspected:** 2026-09-09, clean main `679f5cefd6e626a3c67c1e08a433ab786c944883`; masterplan active target 3.1; current ports/core/storage, Makefile/coverage, product triplet and storage durable learning.
- **External grounding:** Official Go 1.25 time/tzdata and RFC 9562 sections consulted for month normalization, ambiguous-time handling, packaging ownership and UUID layout. No dependency installation or runtime exploration.
- **Confidence review:** Contract and dependency gaps were resolved in the plan; unknown runtime outcomes remain explicit execution gates. No numeric score is offered as evidence of correctness.
- **Documentation audit:** Passed 2026-09-09 through a temporary `planning-audit` Make recipe while loading the canonical Makefile. Checked eight Markdown files and 97 local links/anchors; 22 requirements; seven ordered units with goal, requirements, dependencies, ownership, approach, red tests, verification, failure/recovery and review fields; 91 unique unchecked scenarios with complete requirement mapping; 16 matching issue-register/detail records; balanced fences and consistent tables; product requirement/check-state preservation; unchanged overall phase completion; exactly four newly checked master planning items; documentation-only worktree scope; and clean `git diff --check`. No Makefile target or application test was added. Audit script was temporary, outside the checkout.
- **Code/tests/dependencies/generated output:** No changes. Application tests, make validate, benchmarks, minimum-toolchain tests and cross-builds not run.
- **Hosted/manual/release/publication:** Not run. No commit, push, PR, merge or publish action authorized by this planning request.

## Implementation release gate

All items remain unchecked during planning. Explicit later-phase handoffs are tracked in ISS-016; they do not masquerade as completed release evidence.

- [ ] U1, U2, U3, U6, U7, U4 and U5 implemented in dependency order.
- [ ] Each unit has observed red-first evidence, focused green assertions and make validate on its actual contents.
- [ ] Each unit synchronized and separately committed before the next under implementation/commit authority.
- [ ] All 91 local Feature 003 scenarios executed with exact revision/environment/command/result evidence.
- [ ] Service/dateparse coverage meets 95%; full/race/script gates pass without weakened exemptions.
- [ ] Safe error category and unknown-outcome checks pass; no unaccepted callback-cause loss remains.
- [ ] Real disk WAL/concurrency/rollback/process-recovery integrity scenarios pass.
- [ ] Minimum Go, five CGO-free service test builds, benchmarks and runbook replay recorded.
- [ ] Full ports.TaskService contract is implemented without stubs and consumer handoffs are documented.
- [ ] All Feature 003 runtime findings fixed or explicitly accepted by the user with scoped evidence; remaining unaccepted local blockers: 0.
- [ ] MASTERPLAN and all affected planning/verification/workorder artifacts synchronized with executed evidence.
- [ ] Hosted/native/CLI/TUI/manual/release status reported separately; publication separately authorized if requested.

U1 acceptance: make validate passed with input/base/query seam tests and 16 real failed-rollback cases; [receipt](../verification-evidence/003/u1.json). Public operation/result/concurrency assertions in V07/V09/V11/V12 remain with their implementing units, consistent with the ban on interim facade stubs.

U2 acceptance: deterministic parser and UUIDv7 pass make validate; [receipt](../verification-evidence/003/u2.json). Tests exposed and fixed uppercase weekday dispatch and the valid year-1 midnight sentinel collision. Production tzdata embedding is still a consumer handoff.

U3 acceptance: make validate passed; [receipt](../verification-evidence/003/u3.json). A reproduced manual-leaf reset defect was fixed with explicit child-removal state. Full public move and concurrent mutation proofs remain in U7/U5.
