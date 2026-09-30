---
feature-id: TUSK
plan-source: docs/plans/2026-09-06-001-feat-tusk-modern-task-system-plan.md
verification-plan: docs/verification-plans/2026-09-06-001-feat-tusk-modern-task-system-verification-plan.md
status: Features 002–005 locally accepted; Feature 006 planning and release gates pending
evidence-scope: Planning findings plus linked local feature acceptance receipts
---

# Tusk Modern Task System Issue Workorder

This register accompanies the [product plan](../plans/2026-09-06-001-feat-tusk-modern-task-system-plan.md) and [verification plan](../verification-plans/2026-09-06-001-feat-tusk-modern-task-system-verification-plan.md). It does not supersede the completed core workorder or close any phase's implementation gate.

**Fixed in plan** means the document now decides the behavior; it does not mean implemented or tested. **Open gate/decision** means the named owner must supply the stated evidence before the dependent unit/phase proceeds. Owners are accountable project roles, not claims that a human has accepted an assignment. G1–G4 are the plan's shared gate IDs.

Gate sequencing: Features 002–005 record local acceptance in their own packs; local main records PR #4 merged. G1 is satisfied for Features 002–005 and remains open for Feature 006. Feature 005 U1/U6 close local graph, minimum-Go, CLI/TUI performance and owned Kitty acceptance. G4 native/hosted/distribution proof remains a release gate.

### Current latency acceptance policy (owner-delegated judgment)

The owner authorized best-practice measurement judgment to prevent incidental
measurement noise from blocking useful progress while preserving the speed goal.
This explicitly supersedes the earlier every-sample policy; historical failing
reports below remain failures under their original policy and are not relabeled.
The exact limits below are an engineering decision for Tusk, not a standard
prescribed by the cited sources.

| Fresh-process case | p90 target | p95 guard | p99 guard | Maximum guard |
|---|---:|---:|---:|---:|
| Help/version, clean or invalid configuration | <5 ms | <7.5 ms | <10 ms | <15 ms |
| List/tree/stats/history, human and JSON | <15 ms | <20 ms | <30 ms | <50 ms |

`make bench-cli` collects three complete runs on the declared reference host.
Each case in each run retains five warmups and 100 consecutive measured samples;
every case in every run must pass independently. Percentiles use nearest rank.
Retain min/p50/p90/p95/p99/max, counts at or above the p90 target, all raw durations,
output size and correctness, exit results, exact binary hash, compiler, OS,
filesystem, CPU, governor, power profile and coordinator GOMAXPROCS. No trimming,
subtracting estimated overhead, selecting the best run, or retrying until green.
The maximum guards prevent severe pauses from disappearing behind a percentile.
These are finite-sample acceptance criteria, not population confidence bounds or
hard real-time guarantees. Retained target misses remain visible even on a pass;
no miss is automatically attributed to host noise.

The timer still covers fresh process launch through exit and complete pipe drain;
builds and isolated fixture setup stay outside timing. Use default Codex Bash,
a cached executable/current-schema disk database, balanced power profile and no
concurrent verification workload. Fixtures and all functional requirements are
unchanged. First use, 10,000 tasks, 1 MiB notes, contention and slow output remain
separate observations. Regressions in p90, bounded tails or correctness still block.

Rationale: [Google SRE](https://sre.google/sre-book/service-level-objectives/)
recommends distributions rather than averages alone and discusses why demanding
100% attainment can obstruct useful delivery. [pyperf's system guidance](https://pyperf.readthedocs.io/en/latest/system.html)
explains scheduler/power-related variance and environment metadata. Those sources
support the method; the thresholds above preserve Tusk's fast interactive purpose.


## Issue register

| ID | Gate | Source / lens | Owner | Severity | Status | Next action and retest |
| --- | --- | --- | --- | --- | --- | --- |
| TUSK-ISS-001 | G1 | Architecture/status | Each feature owner | P1 | Open for 006 | Features 002–005 have synchronized, locally accepted packs. Feature 005 completes all eight units and 110 local scenarios; two release scenarios remain with Feature 006, which must complete its own triplet. |
| TUSK-ISS-002 | G2 | Dependency/security | Phase 2 / Phase 6 owners | P1 | Local proof recorded; native pending | Retain accepted Go 1.25.0, modernc v1.58.0/libc v1.75.6 graph. Feature 002 durable evidence owns local compatibility; native/hosted runtime remains Phase 6. No runtime tests were rerun during Feature 003 planning. |
| TUSK-ISS-003 | G3 | Dependency/performance | Phase 4/5 owners | P2 | Closed locally | Feature 005 U1/U6 retain combined graph/minimum-Go proof plus 84/84 CLI and 78/78 TUI case-runs; see u6-acceptance.json. |
| TUSK-ISS-004 | G4 | Release/operations | Phase 6 release maintainer | P2 | Open gate | Establish release candidate and destination/ownership/license/version policy; run hosted/manual matrix before publication. Evidence: Candidate SHA, job URLs, binary hashes, manual terminal record, checksums/notices and actual authorized publication result when released. |
| TUSK-ISS-005 | — | Core contract/coherence | Core/service reviewers | P1 | Fixed in plan | Retain shipped core; update product R5/R6 and map integration tests without reopening completed Phase 1 work. Evidence: Planning source comparison completed; runtime regression evidence remains Phase 3 execution. |
| TUSK-ISS-006 | — | Rollup/product | Service owner | P1 | Fixed in plan | R7–R9 and mutation table define defaults, reopen-before-rollup, explicit reopen precedence and last-child reset. Evidence: Table-driven lifecycle, nested 100%, both policy settings, rollback; not executed. |
| TUSK-ISS-007 | — | Transactions/data integrity | Storage/service owners | P1 | Fixed in plan | KTD4/KTD5 place every mutation read/write under one IMMEDIATE callback with explicit ownership and no nesting. Evidence: Disk/barrier tests and injected failures prove no lost update/partial graph/event set; not executed. |
| TUSK-ISS-008 | — | Storage/test evidence | Storage/test owner | P1 | Fixed in plan | Use uniquely named memory fixtures for semantics and temp disk/process fixtures for WAL/recovery. Evidence: Read actual journal mode, isolate fixtures and execute disk scenarios; not executed. |
| TUSK-ISS-009 | — | Durability/recovery | Storage/reliability owner | P1 | Fixed in plan | KTD9/KTD10 retain configured NORMAL, bounded failures, consistent backup and preservation of sidecars; no automatic destructive recovery. Evidence: Process termination and reopen/integrity plus backup procedure review; no claim this simulates hardware power loss. |
| TUSK-ISS-010 | — | Embedding/generation | Storage/build owner | P1 | Fixed in plan | Use db/embed.go adjacent to migrations, isolated sqlc package, pinned tool, and tested generation targets. Evidence: Compile/generate reproducibility and stale-output negative fixture; not executed. |
| TUSK-ISS-011 | — | Timeline/scope | Storage/service/CLI/TUI owners | P1 | Fixed in plan | Retain original timeline scope with metadata-only task_events and CLI history parity; Feature 002 now supplies the persistence; Feature 003 owns service event attribution and exposure. Evidence: Atomic event lifecycle, no-op dedupe, no retained old user text, cascade deletion and UI/CLI parity; not executed. |
| TUSK-ISS-012 | — | CLI/destruction | CLI/service owners | P1 | Fixed in plan | R20 defines force without implied recursion, default-No, non-TTY rejection and membership recheck. Evidence: Complete TTY/non-TTY/JSON/flag matrix and stale confirmation fixture; not executed. |
| TUSK-ISS-013 | — | Serialization/parity | CLI/API owner | P1 | Fixed in plan | R19–R23 and DTO table define all commands including history, exact shapes, patch intent, exit split and output failure behavior. Evidence: Decode all command outputs, writer-error fixtures and service/TUI parity; not executed. |
| TUSK-ISS-014 | — | Time/filter/metrics | Service/query owner | P2 | Fixed in plan | R13–R17 establish local-day intervals, explicit tokens, DST/calendar policy, deterministic order and retained-task metric; preserve core exclusive bounds. Evidence: Fixed-clock boundary cases and SQL/core result comparison; not executed. |
| TUSK-ISS-015 | — | UI concurrency/conflicts | TUI/service owners | P1 | Fixed in plan | KTD12/KTD14 define base-snapshot conflicts, request generations, one write and committed-result retention after refresh failure. Evidence: Reverse-order messages, two-process edits and save-then-refresh failure; not executed. |
| TUSK-ISS-016 | — | UI purity/accessibility | TUI owner | P1 | Fixed in plan | R24–R26, key table and scenario fixtures cover deep snapshots, narrow layouts, focus and real terminal checks. Evidence: 100 Views in every state, message sequences, manual target-terminal evidence; not executed. |
| TUSK-ISS-017 | — | Filesystem/terminal security | Storage/CLI/TUI owners | P1 | Fixed in plan | Treat path literally, enforce per-file lifecycle controls, sanitize only presentation, retain escaped JSON, never execute/fetch notes. Evidence: Literal special-character paths, unsafe targets, restrictive permission and ESC/OSC cases; not executed. |
| TUSK-ISS-018 | — | Build/evidence integrity | Build/phase owners | P1 | Fixed in plan | Document actual Makefile behavior, label every future target, keep CGO=0 release-only, and retain historical versus current proof. Evidence: Canonical command review completed; future negative script fixtures and candidate CI proof remain unexecuted. |
| TUSK-ISS-019 | — | Identity/validation | Service/API owners | P2 | Fixed in plan | Generate UUIDv7 with injected clock/entropy, preserve opaque core IDs, handle collision, specify Unicode title and normalized tag limits. Evidence: Deterministic ID bit/format tests, collision/entropy rollback, Unicode/UTF-8 boundaries; not executed. |
| TUSK-ISS-020 | — | Performance claims | CLI/TUI performance owners | P1 | Fixed in plan | Retain mandate, define reference fixtures/all-sample thresholds and raw reporting; observed violations stay open even outside normal fixtures. Evidence: G3 environment decision, raw launch-through-exit data, outlier counts, stress and contention results; not executed. |

## Issue details

### TUSK-ISS-001. Architecture/status

- **Found:** Product planning/source review; severity P1. Current G1 remains open for Feature 006.
- **Owner / affected contract:** Each feature owner; R29; all units; TUSK-V73.
- **Evidence / gap:** Features 002–005 have complete, locally accepted triplets. Feature 005 closes eight units and 110 local scenarios; Feature 006 still has an outline and owns the remaining two release scenarios.
- **Decision / next action:** Keep the umbrella requirements-only. Feature 006 planning is next, awaiting instruction.
- **Retest / closure:** Feature 005 u6-acceptance.json binds local runtime, review, performance, terminal and documentation evidence.
- **Blocking boundary:** G1 blocks Feature 006 implementation until its pack exists; G4 native/hosted release proof remains open.

### TUSK-ISS-002. Dependency/security

- **Found:** Product planning/source review on 2026-09-08; severity P1. Current state: local proof recorded by Feature 002; native/hosted release acceptance pending.
- **Owner / affected contract:** Phase 2 and Phase 6 owners; R12/R27/R29; U24/U1–U5/U21; TUSK-V01.
- **Evidence / original gap:** The original Go 1.24 baseline could not use the selected patched driver graph. Feature 002 subsequently implemented and locally validated Go 1.25.0, modernc v1.58.0/libc v1.75.6.
- **Decision / next action:** Retain the accepted graph for Feature 003; Phase 6 supplies target-native/hosted runtime evidence.
- **Retest / closure:** Feature 002 durable evidence contains local minimum-compiler/engine/transaction/cross-build receipts. Feature 003 planning inspected the current manifest without rerunning those tests.
- **Blocking boundary:** No remaining local storage prerequisite blocks service planning; native runtime acceptance still blocks release. No release waiver is recorded.

### TUSK-ISS-003. Dependency/performance

- **Found:** Product planning/source review; severity P2. Closed locally by Feature 005 U1/U6.
- **Owner / affected contract:** Phase 4/5 owners; R24,R28; U11,U15,U16,U20; TUSK-V51,TUSK-V65.
- **Evidence / gap:** Feature 005 retains the accepted graph, full minimum-Go checks, five-target builds and passing CLI/TUI measurement matrices in u6-acceptance.json.
- **Decision / next action:** Retain source/hash-bound compatibility and performance proof; remeasure affected behavior after future production changes.
- **Retest / closure:** Exact dependency graph, minimum-Go builds, v1 API/purity checks and retained CLI/TUI samples; Feature 005 ISS-021/024 own the evidence.
- **Blocking boundary:** No local dependency/performance blocker remains for Feature 005. Native/hosted release proof stays G4.

### TUSK-ISS-004. Release/operations

- **Found:** Planning/source review on 2026-09-08; severity P2; Open gate.
- **Owner / affected contract:** Phase 6 release maintainer; R27, R29; U21–U23; TUSK-V67–TUSK-V73.
- **Evidence / gap:** No hosted matrix, architecture execution, terminal acceptance or distribution metadata exists.
- **Decision / next action:** Establish release candidate and destination/ownership/license/version policy; run hosted/manual matrix before publication.
- **Retest / closure:** Candidate SHA, job URLs, binary hashes, manual terminal record, checksums/notices and actual authorized publication result when released.
- **Blocking boundary:** G4 remains a Phase 6 release gate. Feature 002 compatibility evidence will not substitute for candidate native/hosted/manual acceptance or publication authority.

### TUSK-ISS-005. Core contract/coherence

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** Core/service reviewers; R5, R6; U8; TUSK-V27, TUSK-V28.
- **Evidence / gap:** Original arbitrary-depth and broad status prose contradicted shipped depth 10 and done transition rules.
- **Decision / next action:** Retain shipped core; update product R5/R6 and map integration tests without reopening completed Phase 1 work.
- **Retest / closure:** Planning source comparison completed; runtime regression evidence remains Phase 3 execution.

### TUSK-ISS-006. Rollup/product

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** Service owner; R7–R9; U8; TUSK-V26–TUSK-V31.
- **Evidence / gap:** Optional completion and leaf-reset semantics were unspecified; done-first rollup can hide an incomplete child.
- **Decision / next action:** R7–R9 and mutation table define defaults, reopen-before-rollup, explicit reopen precedence and last-child reset.
- **Retest / closure:** Table-driven lifecycle, nested 100%, both policy settings, rollback; not executed.

### TUSK-ISS-007. Transactions/data integrity

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** Storage/service owners; R10; U4, U5, U8; TUSK-V14–TUSK-V18, TUSK-V31.
- **Evidence / gap:** Original CRUD-only repository cannot make graph edits/rollup/history atomic or prevent opposite concurrent moves.
- **Decision / next action:** KTD4/KTD5 place every mutation read/write under one IMMEDIATE callback with explicit ownership and no nesting.
- **Retest / closure:** Disk/barrier tests and injected failures prove no lost update/partial graph/event set; not executed.

### TUSK-ISS-008. Storage/test evidence

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** Storage/test owner; R12; U3, U5; TUSK-V10, TUSK-V11, TUSK-V16–TUSK-V19.
- **Evidence / gap:** Shared-memory SQLite was presented as proof of WAL, and shared anonymous URI could mix tests.
- **Decision / next action:** Use uniquely named memory fixtures for semantics and temp disk/process fixtures for WAL/recovery.
- **Retest / closure:** Read actual journal mode, isolate fixtures and execute disk scenarios; not executed.

### TUSK-ISS-009. Durability/recovery

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** Storage/reliability owner; R11, R12; U1, U5, U22; TUSK-V03, TUSK-V04, TUSK-V18, TUSK-V70.
- **Evidence / gap:** Research claimed strong durability/no locks while NORMAL permits power-loss rollback and WAL can return busy.
- **Decision / next action:** KTD9/KTD10 retain configured NORMAL, bounded failures, consistent backup and preservation of sidecars; no automatic destructive recovery.
- **Retest / closure:** Process termination and reopen/integrity plus backup procedure review; no claim this simulates hardware power loss.

### TUSK-ISS-010. Embedding/generation

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** Storage/build owner; R11, R29; U1, U2; TUSK-V02, TUSK-V05.
- **Evidence / gap:** Embed file location, generator version versus config format, and generated code ownership were missing.
- **Decision / next action:** Use db/embed.go adjacent to migrations, isolated sqlc package, pinned tool, and tested generation targets.
- **Retest / closure:** Compile/generate reproducibility and stale-output negative fixture; not executed.

### TUSK-ISS-011. Timeline/scope

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** Storage/service/CLI/TUI owners; R18; U1, U8, U9, U13, U18; TUSK-V14, TUSK-V34, TUSK-V59.
- **Evidence / gap:** TUI promised history with no schema/query/service/CLI counterpart.
- **Decision / next action:** Retain original timeline scope with metadata-only task_events and CLI history parity; Feature 002 now supplies the persistence; Feature 003 owns service event attribution and exposure.
- **Retest / closure:** Atomic event lifecycle, no-op dedupe, no retained old user text, cascade deletion and UI/CLI parity; not executed.

### TUSK-ISS-012. CLI/destruction

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** CLI/service owners; R20; U10, U12, U19; TUSK-V37, TUSK-V43, TUSK-V63.
- **Evidence / gap:** Recursive/force/confirmation semantics allowed divergent destructive behavior in scripts and TUI.
- **Decision / next action:** R20 defines force without implied recursion, default-No, non-TTY rejection and membership recheck.
- **Retest / closure:** Complete TTY/non-TTY/JSON/flag matrix and stale confirmation fixture; not executed.

### TUSK-ISS-013. Serialization/parity

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** CLI/API owner; R19, R21, R22; U12–U15; TUSK-V42, TUSK-V47, TUSK-V49.
- **Evidence / gap:** JSON command coverage, field nullability, clear-field edits and durable-action parity were unspecified.
- **Decision / next action:** R19–R23 and DTO table define all commands including history, exact shapes, patch intent, exit split and output failure behavior.
- **Retest / closure:** Decode all command outputs, writer-error fixtures and service/TUI parity; not executed.

### TUSK-ISS-014. Time/filter/metrics

- **Found:** Planning/source review on 2026-09-08; severity P2; Fixed in plan.
- **Owner / affected contract:** Service/query owner; R13–R17; U7, U9; TUSK-V22–TUSK-V24, TUSK-V32, TUSK-V33.
- **Evidence / gap:** Natural date instants, calendar arithmetic, filter boundaries, sort ties and velocity meaning were ambiguous.
- **Decision / next action:** R13–R17 establish local-day intervals, explicit tokens, DST/calendar policy, deterministic order and retained-task metric; preserve core exclusive bounds.
- **Retest / closure:** Fixed-clock boundary cases and SQL/core result comparison; not executed.

### TUSK-ISS-015. UI concurrency/conflicts

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** TUI/service owners; R10, R19, R25; U10, U17, U19; TUSK-V36, TUSK-V38, TUSK-V57, TUSK-V62.
- **Evidence / gap:** Late loads, repeated submit and stale full-row edits could overwrite current state or duplicate operations.
- **Decision / next action:** KTD12/KTD14 define base-snapshot conflicts, request generations, one write and committed-result retention after refresh failure.
- **Retest / closure:** Reverse-order messages, two-process edits and save-then-refresh failure; not executed.

### TUSK-ISS-016. UI purity/accessibility

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** TUI owner; R24–R26; U16–U20; TUSK-V54–TUSK-V66.
- **Evidence / gap:** Render purity assertion omitted internal component state; tiny terminals, modal key leakage and Unicode cells lacked contracts.
- **Decision / next action:** R24–R26, key table and scenario fixtures cover deep snapshots, narrow layouts, focus and real terminal checks.
- **Retest / closure:** 100 Views in every state, message sequences, manual target-terminal evidence; not executed.

### TUSK-ISS-017. Filesystem/terminal security

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** Storage/CLI/TUI owners; R2, R23; U3, U14, U18; TUSK-V09, TUSK-V48, TUSK-V60.
- **Evidence / gap:** Environment path could become DSN options; stored notes could control terminal; filesystem permissions were unspecified.
- **Decision / next action:** Treat path literally, enforce per-file lifecycle controls, sanitize only presentation, retain escaped JSON, never execute/fetch notes.
- **Retest / closure:** Literal special-character paths, unsafe targets, restrictive permission and ESC/OSC cases; not executed.

### TUSK-ISS-018. Build/evidence integrity

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** Build/phase owners; R27–R29; U2, U15, U20–U23; TUSK-V05, TUSK-V67–TUSK-V73.
- **Evidence / gap:** Original docs omitted coverage in validate, conflated CGO-free release and race builds, and used nonexistent focused/generator gates.
- **Decision / next action:** Document actual Makefile behavior, label every future target, keep CGO=0 release-only, and retain historical versus current proof.
- **Retest / closure:** Canonical command review completed; future negative script fixtures and candidate CI proof remain unexecuted.

### TUSK-ISS-019. Identity/validation

- **Found:** Planning/source review on 2026-09-08; severity P2; Fixed in plan.
- **Owner / affected contract:** Service/API owners; R3, R4; U4, U7; TUSK-V12, TUSK-V20, TUSK-V25.
- **Evidence / gap:** ID scheme/generator failure and character/byte semantics were unsettled.
- **Decision / next action:** Generate UUIDv7 with injected clock/entropy, preserve opaque core IDs, handle collision, specify Unicode title and normalized tag limits.
- **Retest / closure:** Deterministic ID bit/format tests, collision/entropy rollback, Unicode/UTF-8 boundaries; not executed.

### TUSK-ISS-020. Performance claims

- **Found:** Planning/source review on 2026-09-08; severity P1; Fixed in plan.
- **Owner / affected contract:** CLI/TUI performance owners; R28; U15, U20; TUSK-V51, TUSK-V65.
- **Evidence / gap:** Unbounded sub-15ms prose lacked workload, process measurement, noise treatment and evidence boundaries.
- **Decision / next action:** Retain mandate, define reference fixtures/all-sample thresholds and raw reporting; observed violations stay open even outside normal fixtures.
- **Retest / closure:** G3 environment decision, raw launch-through-exit data, outlier counts, stress and contention results; not executed.

## Sequential review lenses

Reviewed in the main thread under the repository's sequential-agent tool mapping. These are document-review conclusions, not independent reviewer votes or software sign-offs. Recheck the affected lens whenever a gate changes a contract.

| Lens | Planning status | Concrete correction / remaining gate | Execution retest |
| --- | --- | --- | --- |
| Architecture and sequencing | Reviewed with gates | Product/feature authority, ports and transaction ownership; G1/G2 remain | U1–U10 and phase handoffs |
| Product/scope and agent parity | Reviewed | Preserve timeline and optional completion; expose all durable TUI actions in CLI | V29, V34, V42, V59, V64 |
| Correctness and reliability | Reviewed | Rollup/reopen/move/no-op/commit failure rules | V14–V18, V26–V38 |
| Security/privacy | Reviewed with dependency gate | Patched engine gate, literal paths, restricted files, terminal controls, metadata-only history | V01, V09, V34, V48, V60 |
| Data integrity and migration | Reviewed | Migration atomicity, checksum/version checks, consistent backup, real disk evidence | V02–V04, V14–V19, V70 |
| Performance/concurrency | Reviewed with measurement gate | IMMEDIATE writer, snapshots, finite waits; reference fixtures and actual process timings | G3; V16–V19, V51, V65 |
| CLI/API compatibility | Reviewed | DTOs/nulls/arrays, full command grammar, clear intent, exit/stream split | V39–V53 |
| TUI interaction/accessibility | Reviewed with manual gate | Pure View, request generations, form focus, tiny sizes, Unicode, terminal restoration | V54–V66 |
| Portability/deployment | Reviewed with hosted gate | CGO/race separation, five release targets, Bash/Make prerequisite, data preservation | G4; V67–V72 |
| Test strategy/evidence | Reviewed | 29 requirements mapped to 24 product handoff units and 73 unexecuted scenarios; Feature 002 adds its six-unit execution pack | All scenario execution records |
| Simplicity/maintainability | Reviewed | No plugin framework, global cache, event sourcing, or adapter duplication; small history table serves original scope | Ownership/dependency review at implementation |
| Documentation coherence | Audited | Original product pass: six Markdown files and 38 local links. Feature 002 reconciliation: nine Markdown/MDC files and 55 links; ID/table/fence/whitespace checks passed, implementation checkboxes preserved | Current audit recorded in the Feature 002 workorder |

## Planning evidence record

- Inspected baseline: `6128d312921cecca2a024bc8647d959126822f0b`, branch `main`, initially clean; 2026-09-08.
- Read target, masterplan, governance/rules, concepts, all feature outlines, relevant core implementation/test contracts, Makefile/scripts and research. No source, tests, dependencies or generated code changed.
- Official documentation and selected module manifests inspected; no driver installed, compatibility build or application tests executed.
- Product planning completed with G1–G4 still open. Exact dependency/toolchain compatibility is not asserted from a module fetch.
- Documentation structural audit: **passed**, 2026-09-08. A temporary documentation-only `planning-audit` recipe was added to the Make invocation with `--eval` while loading the canonical Makefile; no repository target or application test was added. Checked 6 Markdown files, 38 local links, balanced fences, table shapes, Markdown whitespace, 29 requirement mappings, 23 complete unit field sets, 73 unique unchecked scenarios, and 20 issue records. Active phase/target/completion and every existing master implementation checkbox were compared to HEAD and preserved. `git diff --check` passed. Existing intentional two-space Markdown line breaks were retained.
- Final confidence review: contract ownership, unit sequencing and scenario coverage are explicit. Corrected a dependency-cycle risk by separating G2/G3 admission decisions from the later runtime/measurement evidence their units produce. Full implementation readiness remains gated by G1–G4; no numerical confidence score substitutes for their evidence.
- Fresh local software validation: **not run**. Historical Phase 001 evidence remains historical. Hosted/manual/release evidence: **not run**.

## Implementation release gate

Feature 002 reconciliation (2026-09-08): G1 is satisfied for its reviewed pack, G2's planning choice is recorded, and new product U24 precedes U1. Updated compatibility ownership, ignored-operation/unknown-commit handling, supplied storage fixture rollups, and planned Make targets in all affected product artifacts. The preceding planning evidence record describes the original 23-unit pass; the latest Feature 002 workorder owns the new audit result. No product scenario or runtime gate is checked by this reconciliation.

Leave every item unchecked during planning.

- [ ] G1 completed for each phase before its implementation.
- [ ] G2 compatible patched engine/toolchain decision recorded and validated.
- [ ] All applicable units implemented with recorded red-first evidence.
- [ ] Focused, disk/process, race and canonical aggregate gates pass.
- [ ] All applicable TUSK scenarios executed with exact revision/evidence.
- [ ] G3 dependency/benchmark evidence complete; latency violations resolved.
- [ ] G4 hosted/manual/architecture and distribution gates complete.
- [ ] All runtime issues fixed; any accepted deferral has an explicit owner and user decision.
- [ ] Remaining unaccepted implementation/release issues: 0.
- [ ] Master checklist and every affected feature triplet synchronized.
- [ ] Release publication separately authorized and its result recorded.

## Feature 002 local handoff — 2026-09-08

The [Feature 002 triplet](../plans/2026-09-06-002-feat-sqlite-storage-and-repository-plan.md) now records six implemented units with per-unit validated commits and 67/68 local scenarios accepted. [Storage operations and Phase 3 obligations](../storage.md) and [durable evidence](../verification-evidence/002/README.md) cover the repository boundary, atomic metadata history, migration refusal, process recovery and benchmarks. Product U24 compatibility evidence is available locally; target-native TUSK-V01 acceptance remains pending. This handoff does not check cross-phase service/CLI/TUI or hosted/native product scenarios. At that handoff MASTERPLAN.md advanced to Feature 003 planning. Its completed planning pack is recorded in the subsequent Feature 003 handoff below.

## Feature 003 planning handoff — 2026-09-09

The [Feature 003 plan](../plans/2026-09-06-003-feat-task-service-engine-plan.md), [verification matrix](../verification-plans/2026-09-06-003-feat-task-service-engine-verification-plan.md), and [workorder](../workorders/2026-09-06-003-feat-task-service-engine-issues-workorder.md) now satisfy G1 for service planning. Product U6 → feature U1, U7 → U2, U8 → U3/U6/U7, U9 → U4, and U10 → U5. The service pack has 22 requirements, seven units and 91 unexecuted scenarios; all 73 product scenario IDs/check states remain unchanged.

TUSK-V20–V38 service coverage is expanded in that matrix. TUSK-V35 filtered TUI presentation and TUSK-V38 post-commit output/refresh presentation still need the owning adapters. The new service workorder separately tracks failed-rollback error-category compatibility (U1), real disk atomicity/recovery and runtime/benchmark proof (U5), and production timezone-data/CLI/TUI/native/hosted handoffs (Features 004–006). No product requirement or runtime/release checkbox closes in this planning update. The Feature 003 workorder owns the current documentation audit; no application tests or make validate ran.


## Feature 004 planning handoff — 2026-09-09

The [Feature 004 plan](../plans/2026-09-06-004-feat-cli-interface-and-scripting-plan.md), [verification matrix](../verification-plans/2026-09-06-004-feat-cli-interface-and-scripting-verification-plan.md) and [workorder](../workorders/2026-09-06-004-feat-cli-interface-and-scripting-issues-workorder.md) now satisfy G1 for CLI planning. Features 002/003 are locally accepted and merged according to MASTERPLAN.md; earlier planning handoffs above are historical.

Product U11→feature U1, U14→U5/U4, U12→U2/U7, U13→U3 and U15→U6. Execution order is product U11→U14→U12→U13→U15 / feature U1→U5→U4→U2→U7→U3→U6. Existing product IDs and all 73 scenario check states are preserved. Feature 004 has 25 requirements, seven units, 91 unchecked scenarios and 25 findings; 20 are fixed in planning and five remain execution/release evidence gates.

TUSK-V39–V53 expand into that matrix, with TUSK-V38 output recovery and TUSK-V73 governance included. G3 CLI dependency choices are recorded in Feature 004 KTD1; U1 still must prove the combined graph/Go 1.25/full executable builds, and U6 must meet the unchanged product performance protocol. V90–V91 retain native/hosted release obligations with Feature 006. G1 stays open for Features 005/006. No application tests, dependency builds or make validate ran, and no runtime/release checkbox closes during planning. MASTERPLAN names Feature 004 U1 as next, awaiting implementation authorization.

## Feature 004 execution handoff — 2026-09-28

Feature 004 has six separately validated unit commits; U6 remains active. Its
[verification matrix](../verification-plans/2026-09-06-004-feat-cli-interface-and-scripting-verification-plan.md)
checks V01–V86 and V88–V89 with [local evidence](../verification-evidence/004/README.md).
Canonical tests, race, coverage, generated checks, minimum Go and five-target
cross-builds pass. Actual Linux process/PTY recovery and visible Kitty checks pass.
V87 reference latency fails on the 1,000-task fixture; ISS-023 blocks U6/Phase 4
acceptance and Phase 5 advancement. No performance contract was weakened.

[CLI](../cli.md) and [service](../service.md) document the current consumer contract.
Feature 005 owns TUI registration, drafts, consent and refresh using the same ports
and explicit outcome handling. Feature 006 owns completion/man pages, native
Windows/macOS and architecture runtime, hosted checks and release artifacts. CLI
V90–V91 and carried storage/service native obligations remain pending. This handoff
does not close cross-phase product scenarios or authorize publication.

### Feature 004 broader optimization follow-up — 2026-09-28

The owner retained 15 ms. Storage/query allocation and sorting changes pass
canonical validation, minimum-Go tests/cross-builds and fresh Kitty checks, but
the all-sample latency gate still fails. Feature 004 U6 and Phase 4 stay open;
no downstream implementation or acceptance is inferred. See the synchronized
[Feature 004 evidence](../verification-evidence/004/u6-broader-checkpoint.json).

## U6 local acceptance — 2026-09-28

Feature 004 U6 and Phase 4 are locally accepted under the owner-delegated distribution
policy above. The [acceptance receipt](../verification-evidence/004/u6-acceptance.json)
records exact source hashes, commands, results and review coverage. This section
supersedes earlier incomplete checkpoints; those reports remain historical evidence.

- `make validate build check-generated`: passed in Codex Bash, including race,
  coverage and schema generation checks. CLI 96.8%, main 95.7%, harness 95.4%.
- `GOTOOLCHAIN=go1.25.0 make test build-cli`: full tests and all five CGO-free
  executable/test target builds passed.
- Three complete reference runs passed all 84 case-runs. Worst query p90
  14.621 ms, p95 16.685 ms, p99 19.125 ms, maximum 23.934 ms. Help/version worst
  p90 2.855 ms, maximum 4.118 ms. All 8,400 measured samples are retained,
  including 40 query samples at or above 15 ms; this is distribution acceptance,
  not an every-invocation guarantee or a statistical population-confidence claim.
- The first distribution run and the default-runtime experiment each failed
  one tree case; neither was discarded or relabeled. The final code change
  replaced repeated graph hash lookups with task indices, reducing workspace
  allocation while preserving ID, parent, cycle, depth and detached-value rules.
  Its allocation test failed at 394,352 bytes before the fix and passed the
  350 KiB limit afterward. The earlier single-thread runtime override was removed.
- Fresh owned Kitty output verifies decimal progress, depth, parent rollup and
  statistics. The temporary database/window was released after inspection.
- Fresh `ce-code-review` completed with no actionable findings. Local personas
  ran inline as required; Composer's served identity was unverified, so no
  independent corroboration is claimed. Peer dispositions are retained.

V01–V89 and ISS-021/022/023/025 are closed locally. The U6 commit contains this
receipt and synchronized acceptance checks. V90–V91 / ISS-024 remain pending
Feature 006 native/hosted release proof. The next target is Feature 005 planning;
its implementation has not started. No push, PR, merge or publication occurred.

## Feature 005 planning handoff — 2026-09-29

The [Feature 005 plan](../plans/2026-09-06-005-feat-interactive-tui-application-plan.md), [verification matrix](../verification-plans/2026-09-06-005-feat-interactive-tui-application-verification-plan.md) and
[workorder](2026-09-06-005-feat-interactive-tui-application-issues-workorder.md) complete G1 for TUI planning: 28 feature requirements, eight
units, 112 unexecuted scenarios and 25 findings/gates (20 corrected in planning,
five execution/release gates). Product U16→feature U1/U7/U2, U17→U3, U18→U4,
U19→U5/U8 and U20→U6. Order: U1 → U7 → U2 → U3 → U4 → U5 → U8 → U6.
Existing product requirement, handoff and scenario IDs/check states are preserved.

Local main at e899491 records Feature 004 PR #4 merged; its local acceptance
receipts remain valid historical evidence, not rerun by this pass. Feature 005
selects its v1 dependency graph and defines prepared-frame purity, single-operation
service ownership, draft/consent conflicts, raw-text preservation and fresh-owner
unknown-outcome readback. G3 remains open for its U1 build proof/U6 measurement;
V01–V110 are local feature obligations, V111–V112 stay with Feature 006 native and
hosted release proof. The TUI source/Makefile targets do not exist yet.

TUSK-V54–V66 now map to that detailed matrix, with TUSK-V35/V38 and governance
covered at the consumer boundary. No product scenario or implementation/release
checkbox closes from planning. No application tests, make validate, benchmarks
or Kitty acceptance ran. MASTERPLAN's next unit is Feature 005 U1 only after
an implementation instruction. Historical handoffs above retain their dated
meaning; this is the current planning handoff.

### Feature 005 runtime compatibility correction — 2026-09-29

Feature 005 U7 exposed an eager global terminal-color query in Bubble Tea
v1.3.10's package initializer. The local dependency copy removes only that
initializer, retains the upstream version/APIs/license, and records source
hashes in `third_party/bubbletea/TUSK-PATCH.json`. Actual child PTY regressions
prove the unsolicited query is absent and ordinary CLI/TUI startup works again.
This amends the selected v1 runtime without a v2 migration or weaker CLI latency
gate. U6 still owns final-candidate distributions; native/hosted proof stays with
Feature 006. See the Feature 005 U7 receipt for execution evidence.

### Feature 005 presentation checkpoint — 2026-09-29

Feature U1/U7/U2 now provide product U16's pure root, owned lifecycle and bounded
layout foundation. The owner selected spacious title/metadata rows after a
Kitty prototype and requested uniform modal/panel backgrounds. The production
surface and actual terminal resize fixes pass canonical, minimum-Go and owned
Kitty checks; see [U2 evidence](../verification-evidence/005/u2.json).
Product TUSK-V54/V55 retain their later form/selection/saving repetitions;
U17–U20 remain pending. Feature U3 is the next execution target. No release,
native-platform or final CLI-distribution acceptance is added by this checkpoint.

### Product U17 navigation handoff — 2026-09-29

Feature 005 U3 supplies forest grouping, navigation, selection identity, literal
search, filters and single-operation refresh; see the synchronized
[U3 receipt](../verification-evidence/005/u3.json). TUSK-V58 passes injected
search/timer, stale snapshot, retry and resize checks. TUSK-V56 has list/filter
and Kitty proof but retains final detail/form keyboard coverage in U4/U5/U6;
TUSK-V57 retains post-mutation/form identity proof in U5/U8. U18 is the current
feature target. No native, hosted, release or complete TUI claim is added.

### Product U18 details handoff — 2026-09-29

Feature 005 U4 supplies wrapped metadata, safe asynchronous Markdown and actual
ordered history; TUSK-V59/TUSK-V60 pass current synthetic, disk and owned Kitty
checks at all target sizes. See the [U4 receipt](../verification-evidence/005/u4.json).
Large/pathological notes use labeled full-content plain text under the explicit
formatting budget; raw storage is preserved. U19 is next. Form identity, writes,
consent/recovery UI and final combined-flow proof remain their designated units.


### Feature 005 U5 handoff — 2026-09-29

Feature 005 U5 now supplies create/edit/move/lifecycle forms with detached Base,
raw preservation, explicit conflict/discard and one-write/readback barriers.
[U5 evidence](../verification-evidence/005/u5.json) records canonical/minimum-Go
and real Kitty checks. TUSK-V62 closes. TUSK-V61/V63 keep their remaining delete
consent portions open for U8. Product U19 and Feature 005 overall acceptance remain
open through U8/U6; native/hosted release evidence remains Feature 006.


### Feature 005 U8 handoff — 2026-09-29

[U8 evidence](../verification-evidence/005/u8.json) closes product TUSK-V61/V63
and the U19 forms/consent implementation. Exact preview consent, fresh-owner
unknown-outcome readback, no replay, explicit acknowledgment and real Kitty
checks pass. U20 / Feature 005 U6 final workflow, performance, documentation and
review remain active; native/hosted release evidence remains Feature 006.

### Feature 005 U6 visual and validation checkpoint — 2026-09-29

The owner approved the refined Spacious production TUI, including padding,
tabs, checkboxes and opaque dialog backgrounds. Feature V101–V105 and V109
pass: workflow, disk/lifecycle, child PTY, owned Kitty, canonical validation,
minimum-Go/five-target builds and documentation. The visual proof is separate
from automated test output.

Product U20 remains open. The final linked CLI matrix passes 74/84 cases after
removing eager syntax/CSS registry initialization from the pinned Markdown
renderer. All failed samples are retained. Final performance and code-review
settlement are required before Phase 5 acceptance and the U6 local commit.
Native macOS/Windows and hosted release proof remain Feature 006 obligations.

### U6 final TUI measurements and open CLI gate (2026-09-29)

All 69 current-binary TUI case-runs pass. Synchronous preparation worst p95 is
11.245 ms and maximum 12.361 ms; pure View remains allocation-free. Three real
child-PTY startup runs retain 300 samples plus 15 warmups: median about 27 ms,
worst run p95 43.259 ms and maximum 48.828 ms, with maximum child RSS 23,620 KiB.
Async 32 KiB Markdown remains an observation (run 3 p95 about 600 ms); 1 MiB
plain fallback about 49 ms. Neither is claimed as synchronous frame preparation.
Feature V107/V108 and product TUSK-V54/V55/V56/V57/V64/V65 now pass. Native
terminal TUSK-V66 and Feature V111/V112 remain explicitly deferred in part/all.

The post-package-update CLI matrix still passes only 79/84: five 1,000-task JSON
list/tree case-runs miss p90, with worst p90 19.317 ms against 15 ms. Help/version
and other query cases pass. The same-host Feature 004 control passes 82/84.
Profiling puts about half of repeated CLI CPU in task reads and about 19% in
storage open; JSON formatting is about 8%. Runtime GC/processor experiments are
retained diagnostics only and do not supply an acceptance substitute. No limits,
fixtures, database safety rules or runtime defaults were changed. V106 and final
U6 acceptance remain open pending a stable-host measurement or a justified fix.

### U6 due-date usability refinement (2026-09-30)

The owner approves the current visual direction and requests a due-date calendar
and accepted-input examples beside the label. This extends the existing U6
form polish (Feature R9/R12/R16/R23/R24; V68/V70/V86/V104), without a new phase.
Keep Spacious rows, shared controls, padding and opaque dialog surfaces. Ctrl+P
on Due opens a month grid; arrows move by day/week, PgUp/PgDn by month, and `t`
jumps to today. Enter copies an ISO day into the draft; Esc leaves the original
text untouched. Typed dates and natural expressions remain available. Examples
and timezone stay visible; selecting a day uses the existing local end-of-day
parser at save. Verify leap/month boundaries, local today, cancellation, focus
trapping, unchanged raw timestamps, resize/pure View, real storage readback and
owned Kitty at 80×24/120×40 in color and plain presentation. These new subcases
pass canonical validation and real Kitty inspection; see the
[calendar receipt](../verification-evidence/005/u6-calendar.md). The owner
approved the calendar on 2026-09-30: “it looks good”; prior UI approval remains. All 78
current-candidate TUI case-runs pass, including nine calendar navigation cases:
calendar preparation worst p95 12.334 ms, maximum 12.968 ms. CLI V106 /
005-ISS-024 and the U6 local commit remain open; the last CLI 79/84 report is
explicitly the pre-calendar candidate, not acceptance of the new binary.

### U6 final measurement context (2026-09-30)

The owner clarified that Zed hosts this Codex terminal session. Keep the editor
running and record its ambient load. As the product latency policy specifies,
run no concurrent verification workload; do not require the owner to close the
session host. After minimum-Go checks finish, quit only the owned Kitty app and
measure the current calendar candidate once through the unchanged three-run
CLI matrix. Preserve the earlier 79/84 report and all new samples. No performance
limit, compiler default, power setting, fixture or storage safety rule changes.


### U6 final JSON correction and acceptance checkpoint (2026-09-30)

The calendar candidate passed 82/84 CLI case-runs: two 1,000-task JSON tree
p90 values were 15.029 and 15.174 ms. A profile-guided formatter correction
preserves exact JSON bytes and lowers isolated encoding median by 18.8%.
Canonical validation, full Go 1.25 tests/five-target builds, and real Kitty
CLI readback pass. The approved TUI/calendar source is unchanged; its retained
78-case TUI matrix remains passing. V28/V29 now have explicit final selection,
draft, refreshing and saving geometry coverage as well.

The corrected binary's full CLI matrix passes 76/84, with query p90 up to
19.110 ms and first-run help/version p90 up to 11.568 ms. Later help/version
runs pass. Both complete reports and all samples are retained; no claim assigns
every miss to host load. See [formatter receipt](../verification-evidence/005/u6-json-formatter.md)
and [current checkpoint](../verification-evidence/005/u6-checkpoint.json).
U6, V106, V110 and 005-ISS-024 remain open. The owner is being asked whether to
retain this local acceptance gate or explicitly hand it to Feature 006 as a
release blocker. No such handoff, threshold change or local completion is
assumed. Zed and the user's other applications remain untouched.

### Feature 005 final local acceptance (2026-09-30)

The owner closed Zed and moved this session to Konsole, explicitly requesting
continued verification. A new full matrix of the unchanged final binary passes
84/84 CLI case-runs (three runs, five warmups and 100 retained samples per case).
Worst query p90/p95/p99/max: 13.873/14.533/19.546/27.451 ms; help/version:
4.592/4.866/6.279/6.297 ms. All original distribution limits pass. The earlier
76/84 report is preserved as u6-logs/cli-latency-before-konsole.json; no failed
sample was removed and no performance gate was waived or handed off.

All 78 TUI measurement case-runs, canonical validation, minimum-Go tests and
five-target builds pass. Source hashes still match those validated and reviewed;
this acceptance adds only evidence and documentation. Owned Kitty inspection
and owner approval cover the Spacious layout, opaque dialogs, padding, tabs,
checkboxes, calendar and due-input examples. The app is reopened in the owned
window against its isolated database; test and timing output stayed in Bash.

[U6 acceptance](../verification-evidence/005/u6-acceptance.json) binds the
candidate hashes, eight unit commit boundaries, red/green receipts, canonical
logs, three completed code reviews, terminal evidence and retained measurements.
V01–V110 and ISS-021–ISS-024 are closed locally. Product U20 is locally accepted;
the non-Linux portion of TUSK-V66 remains open alongside V111–V112/ISS-025.

This containing U6 commit closes Phase 5. Feature 006 planning is next, awaiting
instruction; its native/hosted release checks and publication authority remain
separate. No push, PR, merge or release was performed by this acceptance.
