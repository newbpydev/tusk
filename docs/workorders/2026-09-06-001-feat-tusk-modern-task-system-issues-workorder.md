---
feature-id: TUSK
plan-source: docs/plans/2026-09-06-001-feat-tusk-modern-task-system-plan.md
verification-plan: docs/verification-plans/2026-09-06-001-feat-tusk-modern-task-system-verification-plan.md
status: Open - phase gates pending
evidence-scope: Planning findings only
---

# Tusk Modern Task System Issue Workorder

This register accompanies the [product plan](../plans/2026-09-06-001-feat-tusk-modern-task-system-plan.md) and [verification plan](../verification-plans/2026-09-06-001-feat-tusk-modern-task-system-verification-plan.md). It does not supersede the completed core workorder or close any phase's implementation gate.

**Fixed in plan** means the document now decides the behavior; it does not mean implemented or tested. **Open gate/decision** means the named owner must supply the stated evidence before the dependent unit/phase proceeds. Owners are accountable project roles, not claims that a human has accepted an assignment. G1–G4 are the plan's shared gate IDs.

Gate sequencing: Feature 002 now records local acceptance of G2's runtime and all six units in its [durable evidence](../verification-evidence/002/README.md). Features 003/004 planning G1 is satisfied; Feature 003 is locally accepted and merged. G1 remains open for Features 005/006. G3's dependency/production timezone-data choices precede U11/U16, while U15/U20 supply measurement before phase acceptance. G4 native/hosted/distribution evidence remains a release gate.

## Issue register

| ID | Gate | Source / lens | Owner | Severity | Status | Next action and retest |
| --- | --- | --- | --- | --- | --- | --- |
| TUSK-ISS-001 | G1 | Architecture/status | Each Feature 002–006 owner | P1 | Open for 005–006 | Features 002/003 are locally accepted and merged; Feature 004 has its complete seven-unit, 91-scenario planning pack. Its U1 is next under later implementation authority. The umbrella stays requirements-only; later owners complete their own triplets. |
| TUSK-ISS-002 | G2 | Dependency/security | Phase 2 / Phase 6 owners | P1 | Local proof recorded; native pending | Retain accepted Go 1.25.0, modernc v1.58.0/libc v1.75.6 graph. Feature 002 durable evidence owns local compatibility; native/hosted runtime remains Phase 6. No runtime tests were rerun during Feature 003 planning. |
| TUSK-ISS-003 | G3 | Dependency/performance | Phase 4/5 owners | P2 | Open execution gate | Feature 004 KTD1 selects CLI pins; U1 proves the combined graph/production timezone data and U6 meets the unchanged performance protocol. Feature 005 dependency selection and U20 measurement remain separate. Evidence: module/build logs, raw process samples, reference manifest; not executed. |
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

- **Found:** Planning/source review on 2026-09-08; severity P1; Open for Features 005/006.
- **Owner / affected contract:** Each Feature 002–006 owner; R29; all units; TUSK-V73.
- **Evidence / gap:** Features 005/006 still lack completed triplets; Features 002/003 are locally accepted and Feature 004 has its complete planning pack.
- **Decision / next action:** Keep umbrella requirements-only; G1 is satisfied for Features 002/003/004. MASTERPLAN names Feature 004 U1 as the next implementation unit under later authorization. Apply G1 to Features 005/006.
- **Retest / closure:** Feature 004 has seven units, 25 requirements and 91 mapped scenarios; its workorder records the current planning audit. Feature 002/003 runtime evidence remains in their own packs. Later G1 closures need their own evidence.
- **Blocking boundary:** G1 still blocks Features 005/006 until their packs exist. Feature 004 planning readiness does not grant implementation authority or waive its compatibility/process/latency gates.

### TUSK-ISS-002. Dependency/security

- **Found:** Product planning/source review on 2026-09-08; severity P1. Current state: local proof recorded by Feature 002; native/hosted release acceptance pending.
- **Owner / affected contract:** Phase 2 and Phase 6 owners; R12/R27/R29; U24/U1–U5/U21; TUSK-V01.
- **Evidence / original gap:** The original Go 1.24 baseline could not use the selected patched driver graph. Feature 002 subsequently implemented and locally validated Go 1.25.0, modernc v1.58.0/libc v1.75.6.
- **Decision / next action:** Retain the accepted graph for Feature 003; Phase 6 supplies target-native/hosted runtime evidence.
- **Retest / closure:** Feature 002 durable evidence contains local minimum-compiler/engine/transaction/cross-build receipts. Feature 003 planning inspected the current manifest without rerunning those tests.
- **Blocking boundary:** No remaining local storage prerequisite blocks service planning; native runtime acceptance still blocks release. No release waiver is recorded.

### TUSK-ISS-003. Dependency/performance

- **Found:** Planning/source review on 2026-09-08; severity P2; Open gate.
- **Owner / affected contract:** Phase 4/5 owners; R24, R28; U11, U15, U16, U20; TUSK-V51, TUSK-V65.
- **Evidence / gap:** Feature 004 KTD1 selects the CLI dependencies; combined-graph builds, canonical process benchmarks and reference performance evidence do not yet exist. Feature 005 owns its later UI graph.
- **Decision / next action:** Feature 004 KTD1 pins the CLI source baseline; its U1 proves the graph and production timezone data, then U6 adds benchmark execution. Feature 005 selects its UI graph and product U20 owns its later measurement. Declare the reference environment before either measurement.
- **Retest / closure:** Exact dependency graph, canonical target tests, raw CLI/TUI samples and v1 API proof; close separately for each phase.
- **Blocking boundary:** G3 dependency choices precede the owning CLI/TUI units; their measurements precede phase acceptance. Feature 002's recorded G2 choice does not close either part of G3.

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
