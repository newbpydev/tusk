---
feature-id: "003"
plan-source: docs/plans/2026-09-06-003-feat-task-service-engine-plan.md
surface-profiles: [service-api, library, transactional-persistence, concurrency, portability, documentation]
status: Planned - not executed
evidence-scope: Planning only
---

# Feature 003 Task Service Engine Verification Plan

## Verification contract

Companions: [plan](../plans/2026-09-06-003-feat-task-service-engine-plan.md), [workorder](../workorders/2026-09-06-003-feat-task-service-engine-issues-workorder.md), and [masterplan](../../MASTERPLAN.md). Feature-local R/U/KTD/AE IDs refer to the plan. Product R/U/V IDs are prefixed `product` or `TUSK`.

Verify synchronous service calls, exact input/result/error contracts, transaction-scoped graph/history changes, calendar/identity behavior and read models. The baseline is clean main `679f5cefd6e626a3c67c1e08a433ab786c944883`, inspected 2026-09-09. Existing core/storage acceptance is historical evidence from their own packs. All **91 Feature 003 scenarios below are planned and unchecked**; this document contains no service pass/fail results.

Local evidence consists of deterministic unit assertions, real temporary disk integration, concurrency/fault recovery, minimum-Go compilation, benchmarks and runbook replay. Hosted native operating systems, production timezone embedding, real CLI/TUI behavior and process latency belong to Features 004–006. Cross-compilation is local compatibility evidence only. Browser/mobile/HTTP/auth/rate-limit scenarios do not apply to this internal single-user service.

## Requirement coverage

| Requirement | Units | Scenarios | Evidence tier |
| --- | --- | --- | --- |
| R1 | U1, U4, U5 | 003-V01, 003-V02, 003-V12, 003-V88, 003-V90 | Local/architecture |
| R2 | U1, U4, U7 | 003-V03, 003-V10, 003-V11, 003-V49, 003-V60, 003-V65, 003-V71, 003-V74, 003-V75, 003-V76 | Contract/local |
| R3 | U1, U2, U6 | 003-V04, 003-V05, 003-V06, 003-V07, 003-V08, 003-V09, 003-V24, 003-V44, 003-V48, 003-V69 | Local |
| R4 | U2, U6 | 003-V25, 003-V26, 003-V27, 003-V39, 003-V40, 003-V41 | Local/integration |
| R5 | U2, U4 | 003-V13, 003-V14, 003-V15, 003-V16, 003-V17, 003-V18, 003-V19, 003-V20, 003-V21, 003-V22, 003-V23, 003-V24, 003-V68, 003-V89 | Local/portability |
| R6 | U1, U5 | 003-V09, 003-V11, 003-V42, 003-V63, 003-V76, 003-V81, 003-V82, 003-V83, 003-V84, 003-V85, 003-V88 | Local/disk/recovery |
| R7 | U6 | 003-V39, 003-V40, 003-V41, 003-V42, 003-V43 | Local/disk |
| R8 | U1, U6, U7, U5 | 003-V10, 003-V44, 003-V45, 003-V46, 003-V47, 003-V50, 003-V51, 003-V79 | Local/concurrency |
| R9 | U3, U7, U5 | 003-V38, 003-V56, 003-V57, 003-V58, 003-V78 | Local/disk |
| R10 | U3, U6, U7, U5 | 003-V28, 003-V29, 003-V30, 003-V31, 003-V34, 003-V35, 003-V43, 003-V54, 003-V59, 003-V62, 003-V77 | Local/disk |
| R11 | U7, U5 | 003-V49, 003-V50, 003-V51, 003-V58, 003-V63, 003-V77, 003-V86 | Local/disk |
| R12 | U3, U7 | 003-V31, 003-V32, 003-V33, 003-V52, 003-V53, 003-V54 | Local/integration |
| R13 | U3, U7 | 003-V29, 003-V32, 003-V33, 003-V34, 003-V35, 003-V53, 003-V55 | Local |
| R14 | U1, U6, U7 | 003-V06, 003-V07, 003-V48, 003-V58 | Local |
| R15 | U7, U5 | 003-V60, 003-V61, 003-V62, 003-V64, 003-V80 | Local/disk |
| R16 | U3, U6, U7, U5 | 003-V36, 003-V37, 003-V39, 003-V45, 003-V49, 003-V62, 003-V63, 003-V75, 003-V85, 003-V87 | Local/privacy/disk |
| R17 | U2, U4 | 003-V18, 003-V19, 003-V20, 003-V21, 003-V65, 003-V66, 003-V67, 003-V68, 003-V69, 003-V70 | Local/integration |
| R18 | U4 | 003-V71, 003-V72, 003-V73 | Local/integration |
| R19 | U4 | 003-V74, 003-V75, 003-V76, 003-V85 | Local/disk |
| R20 | U5 | 003-V77, 003-V78, 003-V79, 003-V80, 003-V81, 003-V82, 003-V83, 003-V84, 003-V85, 003-V86, 003-V87 | Disk/concurrency/recovery |
| R21 | U2, U5 | 003-V01, 003-V89, 003-V90, 003-V91 | Local/cross-build/benchmark/docs |
| R22 | All | 003-V88, 003-V90, 003-V91 and every unit's canonical aggregate gate | Local/evidence |

Product trace: TUSK-V20/V21 → V01–V12/V44–V48; TUSK-V22–V25 → V13–V27/V68/V89; TUSK-V26–V31 → V28–V64/V77/V78/V86; TUSK-V32–V35 → V65–V76; TUSK-V36–V38 → V79–V85. Here unprefixed V references mean 003-V. TUSK-V38 output/refresh display and TUSK-V35 TUI filtered-context presentation remain adapter-owned; this pack proves their service handoffs. AE1 maps V77; AE2 V53; AE3 V79; AE4 V80; AE5 V19/V68.

## Scenarios

Each checkbox is one acceptance scenario with table/subtest cases as described. Named test families/owned test paths are in the plan's unit entries. Verification requires assertions on returned errors/results **and** authoritative task/history readback where state can change. Inspecting mocks alone cannot close a disk scenario.

### U1. Contracts, constructor and error seam

- [x] 003-V01 **Normal / no I/O:** Construct with valid repo/functions/location and policy false/true; it calls none of them, opens no storage, and does not close the injected owner.
- [x] 003-V02 **Invalid options:** Nil repository interface, clock, ID function or location returns ErrInvalidServiceOptions without panic/dependency calls. Typed-nil object behavior is outside supported injection.
- [x] 003-V03 **Contract:** Compile consumer fixtures for every planned operation/command/result; optional/clear/consent fields are representable without importing service/SQL into ports. Full concrete conformance is asserted in U4, with no interim method stubs.
- [x] 003-V04 **Metadata boundary:** Empty/whitespace title fails, 255 Unicode runes succeeds, 256 fails; default/invalid priority and exact multiline notes follow core/input rules; invalid UTF-8/NUL fails before admission.
- [x] 003-V05 **ID/tag normalization:** Trim ID/title/tag whitespace; existing opaque non-UUID ID remains addressable; normalize case/#/hyphens and sorted duplicates; invalid/empty tag fails. Inputs and snapshots remain unchanged.
- [x] 003-V06 **Patch shape:** Set+clear due/parent, manual-progress+status and manual-progress+parent intent return ErrInvalidCommand, including equal-value directives. Empty service patch is allowed but still reads/validates target/base.
- [x] 003-V07 **Domain validation:** Invalid status, done→blocked, priority 0 on patch, non-done progress 100 and progress -1/101 return the declared domain errors at their preflight or row-dependent boundary; valid done progress 100 is a no-op.
- [x] 003-V08 **Query shape:** All+Statuses, RootOnly+ParentID and malformed text/enums/tags fail with expected command/domain category. Reversed/equal exclusive date bounds are valid and produce an empty result.
- [x] 003-V09 **Time/cancellation:** Already canceled context returns context cause before dependencies. Zero/out-of-range reference time fails safely; one timestamp is captured per valid mutation/time-dependent query and reused for all changes.
- [x] 003-V10 **Base comparison:** Different editable metadata/ID/incarnation conflicts or fails malformed-shape validation as specified. Equal instants with different locations pass; pointer identity does not matter. Ignore UpdatedAt/CompletedAt and parent-derived progress; compare leaf progress. ABA-to-equal is accepted.
- [x] 003-V11 **No partial values:** Callback/commit/read-cleanup failure after assembling a result returns nil task/list/tree/history or zero stats/delete result. No error result exposes Deleted=true or staged IDs.
- [x] 003-V12 **Isolation and ownership:** Sequential/concurrent calls use per-operation state; returned slices/maps/pointers and supplied command/base slices do not alias. No handle escapes or concurrent handle/repository reentry occurs in a callback.

### U2. Date parsing and identity

- [x] 003-V13 **Word tokens:** Fixed reference in UTC and Sao Paulo covers today/tomorrow/tonight and each weekday abbreviation, including matching today and tonight already past; check exact UTC nanoseconds.
- [x] 003-V14 **Token normalization:** Exterior whitespace and mixed-case word/unit tokens succeed; unsupported full weekdays/extra words/embedded whitespace/empty input fail without fallback.
- [x] 003-V15 **Calendar offsets:** +1d/+3d/+1w/+2w/+1m and leading-zero positive N select expected destination date end; week multiplication is checked before conversion.
- [x] 003-V16 **Month end:** Jan 31 +1m clamps to Feb 28/29; Jan 31 +2m resolves Mar 31, not repeated clamping to Mar 28; leap-year and year rollover cases preserve intended date.
- [x] 003-V17 **Absolute dates:** YYYY-MM-DD strict width, leap-day validity, invalid February dates, month/day zero/overflow and year zero are asserted; no Date normalization turns malformed input into success.
- [x] 003-V18 **RFC3339:** Z and positive/negative offsets, optional 1–9 fractional digits, leap seconds, lowercase t/z, absent offset, comma fraction, excess precision, offset ranges and trailing text enforce the documented grammar; exact valid instant survives UTC conversion.
- [x] 003-V19 **DST normal day:** New York 2026-03-08 and 2026-11-01 DayBounds span 23 and 25 hours; tomorrow/+1d retain civil-day meaning. Covers AE5.
- [x] 003-V20 **Day boundaries:** Due exactly start is included, end-1ns included, exactly end excluded, undated excluded; bounds use local date from parsed instant, not the input offset's calendar label.
- [x] 003-V21 **Offset crossing date:** An explicit offset timestamp landing on a different civil day in injected location selects that local day; original due timestamp parsing remains unchanged.
- [x] 003-V22 **Missing/repeated wall time:** Historical midnight-gap and skipped-date fixtures fail with ErrInvalidDate; next-midnight gap fails bounds; repeated exact wall time selects earliest matching instant. Use real IANA examples and a controlled zone fixture for a repeated 20:00 if necessary.
- [x] 003-V23 **Overflow/range:** Zero, negative, fractional or huge N; overflow in 7*N/month/year arithmetic; UTC year falling outside 1–9999 due to offset conversion; no partial bounds or fallback time. Include final-year day-bound overflow.
- [x] 003-V24 **Bad dependency/text:** Nil location, zero reference and invalid UTF-8/NUL fail safely; fixed-reference inputs are unchanged and parser does not inspect process TZ/environment or wall clock.
- [x] 003-V25 **UUID vector/layout:** Deterministic timestamp/entropy reproduces RFC 9562 vector and canonical lowercase form, exact 48 timestamp bits, version 7 and variant 10; no counter/ordering assertion.
- [x] 003-V26 **UUID failure:** Pre-epoch/out-of-range/zero reference, nil/short/error entropy and malformed injected NewID output fail before DB write; error text does not contain entropy-reader text. No alternate ID or timestamp is generated.
- [x] 003-V27 **Same millisecond/reversal:** Distinct deterministic entropy at same millisecond yields distinct valid IDs; reversed clock still encodes its own time. ID source called once for one create attempt, no automatic collision retry.

### U3. Rollup, staging and event selection

- [x] 003-V28 **Floor and depths:** Empty affected set, root leaf, one-level parent and depth-10 chains; children 100/0/0 yield 33, and intermediate flooring propagates exactly to each ancestor.
- [x] 003-V29 **Nested complete progress:** Non-done intermediate parents at 100 contribute 100 upward; this alone never auto-completes their parents because their statuses are open.
- [x] 003-V30 **Last child:** Removing/moving/deleting final child resets open parent to 0 and keeps done parent 100; previous manual progress is not restored. A remaining sibling preserves average semantics.
- [x] 003-V31 **Reopen before rollup:** A done ancestor receiving a new/reopened open child becomes in-progress, clears completion and recalculates; done-first CalculateProgress cannot mask incomplete work.
- [x] 003-V32 **Policy pair:** Identical final nonempty/all-done child fixtures complete parents only with policy true; policy false leaves open parent at 100. Empty children never auto-complete.
- [x] 003-V33 **Explicit open precedence:** Explicitly opened target remains todo/in-progress at 100 with all children done and policy true; done higher ancestors reopen even if the open target still contributes 100.
- [x] 003-V34 **Shared ancestor:** Move between branches with a common grandparent, and ancestor/descendant-related old/new parents; process final deepest-first union, shared grandparent once after both branches, no transient extra event.
- [x] 003-V35 **Status-only propagation:** Child changes done↔open while remaining at progress 100; higher policy/reopen decisions still run despite unchanged numeric progress.
- [x] 003-V36 **Net diffs/no-op:** Helpers may set timestamps internally, but equal canonical final fields produce no row write/event. Changed rows use one operation time; discarded/intermediate fields never reach history.
- [x] 003-V37 **Event attribution:** Exercise create, metadata, move, status, manual-progress and rollup, including compound patches. Assert exact sorted fields, fixed category/task ordering, Sequence=0 on append and no duplicated field across events.
- [x] 003-V38 **Invalid graph/cancellation:** Cyclic/missing-parent/depth/invalid-record fixtures and cancellation during traversal fail without partial results or infinite traversal; no clamped corrupt data is saved.

### U6. Create and patch primitives

- [x] 003-V39 **Create default:** Create root with 255-rune title, notes, tags and due; verify canonical UUID, todo/medium/0, null completion, exact metadata and one create event with all nine allowed fields.
- [x] 003-V40 **Parent/defaults:** Create child under open and done parents at admissible depth; defaults hold, ancestor state/progress/history changes share timestamp and transaction; no implicit parent creation.
- [x] 003-V41 **Failure before persistence:** Entropy failure, invalid generated ID, duplicate storage ID, invalid date and missing parent return their causes with no created row/history; count ID calls to forbid retries.
- [x] 003-V42 **Create atomicity:** Fail Create, each ancestor Update, each AppendEvent or outer commit after staged create; known rollback yields old graph/history; unknown outcome never returns task success.
- [x] 003-V43 **Create affects chain only:** Add incomplete child to previously complete hierarchy and preserve unrelated branch values/events; root creation has no ancestor writes.
- [x] 003-V44 **Set/omit/clear:** Omitted title/notes/priority/tags/due/parent preserved; explicit notes empty/tags empty/due clear/parent clear produce intended values. Repeated semantically equivalent tags and UTC instants count as equal.
- [x] 003-V45 **Metadata no-op:** Empty patch and supplied identical values return current task with unchanged UpdatedAt/events; metadata-only edit never auto-completes/reconciles unrelated graph or existing parent progress.
- [x] 003-V46 **Base conflict:** Change one compared field between Base capture and patch, then reject atomically; valid no-base patch reads latest values and never overwrites omitted concurrent fields.
- [x] 003-V47 **Missing/incarnation:** Missing patch target fails, mismatched Base ID is malformed, same ID with different CreatedAt conflicts; no replacement task is overwritten.
- [x] 003-V48 **Manual progress:** Open leaf 0/99 succeeds; invalid range, any parent manual intent (even equal value) and progress with status/parent intent fail; done leaf only accepts equal 100 as no-op.

### U7. Complete, reopen, move and delete

- [x] 003-V49 **Complete subtree:** Complete a three-level target with mixed statuses; all non-done nodes become done with one time, existing done nodes retain times, ancestors roll up and event fields reflect net changes.
- [x] 003-V50 **Satisfied complete:** Repeating complete on a fully done subtree is a no-op, even under a changed clock/policy; Base validation still occurs.
- [x] 003-V51 **Done target/open descendant:** Seed valid rows representing a done target with an open descendant; explicit completion visits descendants and completes them rather than returning early on target status.
- [x] 003-V52 **Leaf reopen/state machine:** Done leaf→todo/in-progress resets progress0/completion nil; open→same is no-op, open-state changes preserve manual progress; done→blocked and ReopenTask blocked/zero targets fail.
- [x] 003-V53 **Parent reopen:** Preserve done descendants, keep explicitly reopened parent open at calculated 100 with policy true/false, reopen done ancestors, and repeat without duplicate changes. Covers AE2.
- [x] 003-V54 **Child reopen:** Reopen an interior child of a completed depth-10 chain; done ancestors become in-progress bottom-up, siblings remain unchanged and no numeric-unchanged shortcut hides the state change.
- [x] 003-V55 **Auto completion cascade:** Complete final incomplete leaf with policy true; all-done direct children cause upward completion with consistent timestamp; policy false and open-100 intermediate child cases do not.
- [x] 003-V56 **Cycle/existence:** Self-parent matches both cycle sentinels; two-node/deep-cycle/missing-parent moves fail and preserve graph/events; valid root promotion succeeds.
- [x] 003-V57 **Depth:** Move leaf to depth10 succeeds and depth11 fails; moving a subtree includes descendant height. Promotion rebases depth without changing descendants' parent IDs.
- [x] 003-V58 **Compound move/status:** Move first then complete/reopen/open-status patch; only target subtree changes status, both chains reflect final membership, target explicit open wins, progress+status/parent is refused.
- [x] 003-V59 **Move no-op/old chain:** Same parent is no-op unless another field changes; disjoint/shared old/new chains retain final averages, last-child reset and no duplicate shared-ancestor event.
- [x] 003-V60 **Delete preview:** One read snapshot returns full sorted unique subtree IDs and detached target including done descendants; leaf gives one ID; missing target fails and errors return no preview values.
- [x] 003-V61 **Consent/recursion matrix:** Nonforce without Expected returns confirmation-required; malformed preview or Force+Expected invalid; nonrecursive parent returns children-present even when forced; confirmed leaf and forced recursive subtree succeed.
- [x] 003-V62 **Delete integrity:** Exact deleted IDs/count match authoritative subtree; task events cascade, unrelated events survive, old ancestors recompute, final-child reset obeys status, no delete tombstone is created.
- [x] 003-V63 **Failure points:** Table-inject each descendant/ancestor write, event append and delete error; graph/history wholly roll back. Mismatch between validated membership and delete result aborts. Suppress staged target/delete success.
- [x] 003-V64 **Stale consent:** Membership addition/removal/move and target metadata/incarnation change fail equality check; unrelated-task or derived parent-progress changes alone do not conflict. Force executes current scope, preserving Recursive requirement.

### U4. Read surface and facade

- [x] 003-V65 **List statuses:** Default excludes done, All includes every status, explicit statuses select exactly requested states; empty results are non-nil and input filters stay unchanged.
- [x] 003-V66 **Filter parity:** Status/priority OR, tags/fields AND, root/parent, title+notes literal Unicode search, percent/underscore/quotes and nil dates match a core.FilterTasks oracle.
- [x] 003-V67 **Order:** Urgent/high/medium/low, earliest due with nil last, CreatedAt and ID ties; no custom ordering or mutable package-global default-order slice.
- [x] 003-V68 **Due-day intersection:** Combine all/default/status/search/parent with local-day predicate; assert start included/end excluded and due==now is not overdue; no exclusive-bound conversion loses midnight. Covers AE5.
- [x] 003-V69 **Query invalid/empty range:** Invalid enum/tag/UTF-8/NUL and conflicting shape rejected before read; equal/reversed core bounds yield empty; bad Due expression returns date error without admission.
- [x] 003-V70 **Detached reads:** Mutate returned task/tag/date/filter copies and reread; storage and subsequent results unchanged. Caller-owned query slices are never sorted/normalized in place.
- [x] 003-V71 **Forest:** Empty, single root, multiple roots and depth10; includes done descendants; independent sibling ordering; no filtering creates missing parents.
- [x] 003-V72 **Selected subtree:** Select non-root, return one node at Depth1 with descendants rebased; root task keeps original ParentID. Stored graph and later forest depths are unchanged.
- [x] 003-V73 **Tree invalid:** Missing selected ID, orphan, duplicate, cycle and depth overflow produce errors with no partial forest; never detach parent IDs to conceal corruption.
- [x] 003-V74 **Statistics:** Empty yields zeros/four status keys; mixed nested tasks count parents, correct floor percent, strict overdue and (now-168h,now] completion window. Boundary completions, future completion, reopen and delete alter retained metric exactly.
- [x] 003-V75 **History:** Existing no-event task returns empty; nonexistent fails; every event kind ordered by persisted sequence, equal timestamps remain sequence-ordered, changed-field slices detached, deleted timelines absent.
- [x] 003-V76 **Read failure:** Inject cancellation, decode/corruption and callback cleanup failure after reads; all result shapes suppressed. Complete concrete service satisfies every inbound interface method without opening extra snapshots.

### U5. Disk WAL, recovery and handoff acceptance

- [x] 003-V77 **Concurrent siblings:** Two independently opened repositories complete distinct siblings with barriers; final children/ancestor values contain both updates, exact committed event sets and no race. Covers AE1.
- [x] 003-V78 **Opposite moves:** Concurrent A-under-B and B-under-A through separate owners; at most one succeeds, other gets cycle error, resulting forest valid and rejected mutation has no events. Include descendant-height race variant.
- [x] 003-V79 **Stale patch:** Two owners read/edit same task with/without Base; base conflict preserves newer row/draft, no-base disjoint patch preserves omitted fields, unrelated tasks do not falsely conflict. Covers AE3.
- [x] 003-V80 **Stale preview:** Pause after read-only preview, mutate membership/metadata through second owner, then nonforce delete fails; force still needs Recursive. Covers AE4.
- [x] 003-V81 **Writer admission cancel/busy:** Hold independent writer lock, invoke service with short deadline and uncancelled five-second budget, assert context/ErrBusy respectively; release lock and new explicit operation succeeds. No callback replay or event created by rejected call.
- [x] 003-V82 **Statement/event rollback:** Decorate real writer to return error before each Create/Update/Delete/AppendEvent index; inspect full graph/history from a new snapshot and confirm old state. Inject after a real statement as well to prove transaction rollback.
- [x] 003-V83 **Unknown outcomes:** Before-commit and after-real-commit fault modes return ports.TransactionError while storing wholly old or new state respectively; assert errors.As plus all safe causes, nil result, callback exactly once, then close/reopen/readback without retry.
- [x] 003-V84 **Known commit then cancel:** Cancel only after real WithWrite success; service returns committed task/delete result. A later explicit read failure is a separate operation and never replays the write. CLI output behavior remains Feature 004 proof.
- [x] 003-V85 **Snapshot/read cleanup:** Hold one service query callback snapshot while another writer commits a multi-task/history change; assembled result is wholly old or new, not mixed. Read cleanup error suppresses that result; fresh owner reads complete current data.
- [x] 003-V86 **Process termination:** Child process runs service subtree completion and is killed/reaped at before-first-write, after-some-writes and after-acknowledged-commit barriers. Fresh storage open shows complete old/new graph/history with integrity_check=ok and zero FK violations. No WAL/SHM deletion; no hardware-power-loss claim.
- [x] 003-V87 **Privacy and independence:** Notes containing Unicode, ESC/OSC-like content, quotes and private markers persist exactly when allowed; events/errors contain no note/title values or raw driver/entropy text. All DB fixtures are temporary and unrelated roots/owners remain intact.

### Cross-unit compatibility and evidence

- [x] 003-V88 **Safe cause catalog and canonical gates:** Every new immutable port error survives failed rollback individually/joined with context/corruption while private wrappers are redacted and unknown outcome remains detectable. Full/race/coverage/script gates pass, service/dateparse >=95%; no exemption weakening. U1 establishes catalog cases, U5 reruns aggregate acceptance.
- [x] 003-V89 **Portable date fixtures:** Named IANA tests work with test-embedded tzdata; parser uses supplied location and no global TZ mutation. Record Feature 004 production main-package embedding as pending, not proven by unit tests.
- [x] 003-V90 **Compatibility/runbook:** Minimum Go1.25 full suite passes; new build-service target compiles service/tests for Linux amd64/arm64, macOS amd64/arm64, Windows amd64 with CGO0. Replay service lifecycle/uncertain-outcome instructions using temporary DBs; record exact revisions and verify docs/consumer handoffs. Native/hosted targets remain pending.
- [x] 003-V91 **Resource/performance:** New bench-service target measures warm Get/List/Tree/Stats/History and create/complete/reopen/move/delete at 0/100/1000/10000 tasks where applicable, including depth10/wide trees. Record time/allocations, query counts, snapshot lifetime and host details; fixture setup/reset outside timed sections. No global cache, full-DB mutation scan or query-count growth from per-child redundant ancestor reads. Service timings do not close CLI latency gates.

## Commands and environments

Run commands from repository root. Names are future test contracts, not evidence that files exist or tests passed. The current Makefile does not honor PKG/RUN/TEST filter variables.

| Tier | Command or method | Environment | Evidence required |
| --- | --- | --- | --- |
| Unit/red-first | `make test-unit` | Local; named new tests within short suite | Exact initial failure and later green output, owning test path/case |
| Integration | `make test` | Local disposable disk DBs/processes | Graph/event before-after snapshots, fault stage and exit result |
| Concurrency | `make race` | Race-capable host, CGO/compiler available | Channel/barrier schedule and zero races; disk WAL asserted |
| Unit/phase gate | `make validate` | Exact unit/candidate contents | fmt/vet/full/race/coverage/script output and status |
| Minimum runtime | `GOTOOLCHAIN=go1.25.0 make test build-service` | Local, toolchain available; build-service added in U5 | Go version, full suite result, five-target test/binary compile |
| Performance | `make bench-service` | Added/tested by U5 | Raw benchmark output, fixture/host/pin/revision details |
| Doc audit | Temporary `planning-audit` recipe via canonical Makefile | Planning only | Local links/anchors, matrix/IDs/status/fields/whitespace |
| Hosted/native | Owning Phase 6 candidate matrix | Native Linux/macOS/Windows target hosts | Job URLs/revision/architecture runtime; pending here |
| Consumer/manual | Feature 004/005 workflows | CLI subprocesses/real terminal | Date packaging, consent, DTO/exit/refresh/latency; pending here |

Do not rerun expensive gates during planning or add tests mirroring prose formatting. At execution, run the canonical containing suite to observe Red, apply the narrow implementation and run focused/aggregate gates appropriate to the unit. Only broaden/repeat after changes or unresolved failures justify it.

## Fixtures and failure injection

Use immutable fixed times and per-test deterministic entropy. Inject options rather than change process-global clock/location. Reader/clock functions used concurrently must synchronize their own test state. Keep all task IDs explicit in graph fixtures; generated IDs use deterministic UUIDv7 only in creation tests. Assert core setters' persisted values, not only CalculateProgress output.

Unit doubles model WithRead/WithWrite transaction admission with cloned state and commit-on-success; they must discard every staged task/event on callback failure. Writer handles reject escape, concurrent use and repository reentry. Prefer explicit fault index/kind fields and output-state assertions to strict incidental SQL/call-order expectations. Return copied slices. Do not duplicate SQL/entire storage semantics in the fake.

Disk integration uses storage.Open with an explicit t.TempDir path. The public disk factory is sufficient; do not export storage's private connectors or use its internal test functions across packages. Real writer decorators can intercept statements and propagate errors through the actual callback. An outer repository decorator can call the real callback/commit and then substitute an unknown outcome for service no-replay/readback tests; this is a **simulated lost acknowledgment**, not an injected physical driver Commit failure. Storage's existing tests own the physical commit/rollback fault proof. Process children synchronize through explicit files/pipes with deadlines and are always reaped; no sleeps as ordering assumptions.

For corrupt records use a test-owned temporary database connection after valid setup, with narrowly disabled constraints only for the intended fixture, restoring configuration/closing before service reads. For service unknown-outcome propagation, use declared ports.TransactionError values. Never print real user data, load the default user database, unlink sidecars, or weaken production validation for fixture convenience.

Benchmarks mirror storage fixture scale: 128-byte notes, three tags, depth-ten chains plus wide branches. Time warm service work separately from storage open. Reset mutation fixtures between iterations outside timing; validate fixture semantics outside measured loops. Counts/allocations characterize scale without arbitrary new latency thresholds. Any reproducible unexplained regression or redundant full-database work becomes an issue before phase acceptance.

## Execution record

The 2026-09-09 lfg intake rechecked the implementation contract against clean revision `f2fd0e35f16edb78ef5e57c0f4a68ffa7a1b9fa0`. Local sequential coherence, feasibility, scope and adversarial review found no blocking contract gaps. Additional Claude review produced no usable result (HTTP 401, expired OAuth token); independent corroboration is unavailable. This entry is document evidence only; all runtime scenarios below remain pending until executed.

Planning does not fill implementation pass/fail values. Create `docs/verification-evidence/003/` during execution with an index and receipts containing scenario IDs, UTC date, exact commit plus worktree/diff identity for precommit Red/Green, Go/OS/architecture, command, exit code, expected/actual, raw evidence path and unresolved limitation. Attach each unit's validation to its eventual commit without claiming a precommit run ran on a commit that did not yet exist.

| Scope | Current evidence |
| --- | --- |
| Planning/source inspection | 2026-09-09; baseline SHA above; structural result in workorder |
| Feature 003 red/green/integration/race/coverage | Passed; unit receipts and U5 aggregate, service 95.8%, parser 98.4% |
| Feature 003 benchmark/minimum-Go/cross-build | Recorded/passed; U5 raw logs and source hashes |
| Hosted/native/CLI/TUI/manual/release | Not run; later feature owners |

Every executed scenario records its own result. A broad green suite does not close a scenario without verifying that its described assertions ran. Leave all implementation/acceptance boxes unchecked until that evidence exists.

U1 execution: [receipt and raw logs](../verification-evidence/003/u1.json). make validate passed; V06/V08 have helper proof but retain public behavior checks for U6/U4, V07/V09/V11/V12 retain orchestration checks, and V88 retains U5 aggregate acceptance. These open boxes do not imply unimplemented U1 method stubs.

U2 execution: [receipt](../verification-evidence/003/u2.json), make validate passed. V20 predicate membership is retained for U4, while parser day bounds are proven. Year-1 midnight has its own regression.

U3 execution: [receipt](../verification-evidence/003/u3.json), make validate passed. Shared-ancestor and graph-fault helpers have proof; full related-parent moves and public graph validation extend V34/V38 in U7/U5. Untouched manual leaf progress has regression coverage.

U6 execution: [receipt](../verification-evidence/003/u6.json), make validate passed. Initial real-disk creation/patch/rollback cases are present; expanded ancestor-chain, per-index injection and public patch combinations remain in U7/U5. Cached child/parent identity collisions have red-first regression evidence.

U7 execution: [receipt](../verification-evidence/003/u7.json), make validate passed. Every real write/event position fails atomically before/after the statement across create/complete/reopen/move/delete. U5 retains extended concurrency/depth/consent variants.

U4 execution: [receipt](../verification-evidence/003/u4.json), make validate passed. Complete port conformance, detached snapshot reads, selected-tree depth and history are verified. Remaining expanded query/metric cases are explicit U5 aggregate work.

U5 execution: [receipt](../verification-evidence/003/u5.json) maps all 91 scenarios to tests; final make validate and Go 1.25 full suite/five-target builds pass. Additional tests characterize existing behavior; only missing canonical build/benchmark targets required a red-first implementation. Port query counts and transaction lifetime are recorded at 0/100/1000/10000 tasks; these are not SQL VM traces or statistical latency guarantees. Native/hosted/CLI/TUI obligations remain in the runbook. Earlier unit notes describe their then-pending scope, now covered by aggregate acceptance.

Publication review: sequential reuse/quality/efficiency and risk-based code review completed at `d714ea6` with no actionable findings. [Review record](../verification-evidence/003/publication-review.json); [implementation return](../verification-evidence/003/work-return.json). Browser QA is inapplicable because this feature has no browser routes. Bounded hosted feedback monitoring follows publication; merge remains user-owned.
