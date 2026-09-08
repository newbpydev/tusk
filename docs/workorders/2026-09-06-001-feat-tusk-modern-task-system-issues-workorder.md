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

Gate sequencing: G2's compatibility decision must precede U1; U1's first red/green compatibility fixture then supplies runtime proof before production schema code. G3's dependency decisions precede U11/U16, while U15/U20 close its measurement portions before phase acceptance. G4 distribution decisions precede packaging; hosted/manual/publication evidence follows the candidate artifact. These gates do not require a later unit's evidence before the unit producing it can begin.

## Issue register

| ID | Gate | Source / lens | Owner | Severity | Status | Next action and retest |
| --- | --- | --- | --- | --- | --- | --- |
| TUSK-ISS-001 | G1 | Architecture/status | Each Feature 002–006 owner | P1 | Open gate | Keep umbrella requirements-only; reconcile and complete only the active feature triplet before moving its master pointer. Evidence: All applicable R/U/scenario mappings, complete phase units and synchronized checklist; first closure is Phase 2 target 2.1. |
| TUSK-ISS-002 | G2 | Dependency/security | Phase 2 maintainer | P1 | Open decision | Resolve minimum Go/toolchain policy before installing dependencies. Recommended: deliberate minimum raise plus pinned patched modernc/libc; otherwise research a patched Go 1.24-compatible driver. Evidence: Pinned manifests and supported targets, engine version including WAL fix, actual canonical build/integration proof after authorization. Do not infer no alternative exists. |
| TUSK-ISS-003 | G3 | Dependency/performance | Phase 4/5 owners | P2 | Open gate | Pin compatible Cobra/Charm/calendar zone dependencies and implement named Make targets in U15/U20, with the reference environment declared before measurement. Evidence: Exact dependency graph, canonical target tests, raw CLI/TUI samples and v1 API proof; close separately for each phase. |
| TUSK-ISS-004 | G4 | Release/operations | Phase 6 release maintainer | P2 | Open gate | Establish release candidate and destination/ownership/license/version policy; run hosted/manual matrix before publication. Evidence: Candidate SHA, job URLs, binary hashes, manual terminal record, checksums/notices and actual authorized publication result when released. |
| TUSK-ISS-005 | — | Core contract/coherence | Core/service reviewers | P1 | Fixed in plan | Retain shipped core; update product R5/R6 and map integration tests without reopening completed Phase 1 work. Evidence: Planning source comparison completed; runtime regression evidence remains Phase 3 execution. |
| TUSK-ISS-006 | — | Rollup/product | Service owner | P1 | Fixed in plan | R7–R9 and mutation table define defaults, reopen-before-rollup, explicit reopen precedence and last-child reset. Evidence: Table-driven lifecycle, nested 100%, both policy settings, rollback; not executed. |
| TUSK-ISS-007 | — | Transactions/data integrity | Storage/service owners | P1 | Fixed in plan | KTD4/KTD5 place every mutation read/write under one IMMEDIATE callback with explicit ownership and no nesting. Evidence: Disk/barrier tests and injected failures prove no lost update/partial graph/event set; not executed. |
| TUSK-ISS-008 | — | Storage/test evidence | Storage/test owner | P1 | Fixed in plan | Use uniquely named memory fixtures for semantics and temp disk/process fixtures for WAL/recovery. Evidence: Read actual journal mode, isolate fixtures and execute disk scenarios; not executed. |
| TUSK-ISS-009 | — | Durability/recovery | Storage/reliability owner | P1 | Fixed in plan | KTD9/KTD10 retain configured NORMAL, bounded failures, consistent backup and preservation of sidecars; no automatic destructive recovery. Evidence: Process termination and reopen/integrity plus backup procedure review; no claim this simulates hardware power loss. |
| TUSK-ISS-010 | — | Embedding/generation | Storage/build owner | P1 | Fixed in plan | Use db/embed.go adjacent to migrations, isolated sqlc package, pinned tool, and tested generation targets. Evidence: Compile/generate reproducibility and stale-output negative fixture; not executed. |
| TUSK-ISS-011 | — | Timeline/scope | Storage/service/CLI/TUI owners | P1 | Fixed in plan | Retain original timeline scope with metadata-only task_events and CLI history parity; G1 requires Feature 002 schema reconciliation. Evidence: Atomic event lifecycle, no-op dedupe, no retained old user text, cascade deletion and UI/CLI parity; not executed. |
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

- **Found:** Planning/source review on 2026-09-08; severity P1; Open gate.
- **Owner / affected contract:** Each Feature 002–006 owner; R29; all units; TUSK-V73.
- **Evidence / gap:** Feature stubs claim readiness without their own completed triplets.
- **Decision / next action:** Keep umbrella requirements-only; reconcile and complete only the active feature triplet before moving its master pointer.
- **Retest / closure:** All applicable R/U/scenario mappings, complete phase units and synchronized checklist; first closure is Phase 2 target 2.1.
- **Blocking boundary:** G1; revisit when the responsible phase enters planning, before dependent implementation or acceptance. G2 needs the maintainer's compatibility decision; other gates need the identified phase artifacts/evidence. No waiver or user acceptance of an unfulfilled gate is recorded.

### TUSK-ISS-002. Dependency/security

- **Found:** Planning/source review on 2026-09-08; severity P1; Open decision.
- **Owner / affected contract:** Phase 2 maintainer; R12, R29; U1–U5; TUSK-V01.
- **Evidence / gap:** Current Go 1.24 module conflicts with inspected patched modernc candidates requiring Go 1.25; v1.46.1 embeds affected SQLite 3.51.2.
- **Decision / next action:** Resolve minimum Go/toolchain policy before installing dependencies. Recommended: deliberate minimum raise plus pinned patched modernc/libc; otherwise research a patched Go 1.24-compatible driver.
- **Retest / closure:** Pinned manifests and supported targets, engine version including WAL fix, actual canonical build/integration proof after authorization. Do not infer no alternative exists.
- **Blocking boundary:** G2; revisit when the responsible phase enters planning, before dependent implementation or acceptance. G2 needs the maintainer's compatibility decision; other gates need the identified phase artifacts/evidence. No waiver or user acceptance of an unfulfilled gate is recorded.

### TUSK-ISS-003. Dependency/performance

- **Found:** Planning/source review on 2026-09-08; severity P2; Open gate.
- **Owner / affected contract:** Phase 4/5 owners; R24, R28; U11, U15, U16, U20; TUSK-V51, TUSK-V65.
- **Evidence / gap:** CLI/TUI combined dependencies, canonical process benchmarks and reference performance environment do not yet exist.
- **Decision / next action:** Pin compatible Cobra/Charm/calendar zone dependencies and implement named Make targets in U15/U20, with the reference environment declared before measurement.
- **Retest / closure:** Exact dependency graph, canonical target tests, raw CLI/TUI samples and v1 API proof; close separately for each phase.
- **Blocking boundary:** G3; revisit when the responsible phase enters planning, before dependent implementation or acceptance. G2 needs the maintainer's compatibility decision; other gates need the identified phase artifacts/evidence. No waiver or user acceptance of an unfulfilled gate is recorded.

### TUSK-ISS-004. Release/operations

- **Found:** Planning/source review on 2026-09-08; severity P2; Open gate.
- **Owner / affected contract:** Phase 6 release maintainer; R27, R29; U21–U23; TUSK-V67–TUSK-V73.
- **Evidence / gap:** No hosted matrix, architecture execution, terminal acceptance or distribution metadata exists.
- **Decision / next action:** Establish release candidate and destination/ownership/license/version policy; run hosted/manual matrix before publication.
- **Retest / closure:** Candidate SHA, job URLs, binary hashes, manual terminal record, checksums/notices and actual authorized publication result when released.
- **Blocking boundary:** G4; revisit when the responsible phase enters planning, before dependent implementation or acceptance. G2 needs the maintainer's compatibility decision; other gates need the identified phase artifacts/evidence. No waiver or user acceptance of an unfulfilled gate is recorded.

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
- **Decision / next action:** Retain original timeline scope with metadata-only task_events and CLI history parity; G1 requires Feature 002 schema reconciliation.
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
| Test strategy/evidence | Reviewed | 29 requirements mapped to 23 units and 73 unexecuted scenarios; no proxy proof | All scenario execution records |
| Simplicity/maintainability | Reviewed | No plugin framework, global cache, event sourcing, or adapter duplication; small history table serves original scope | Ownership/dependency review at implementation |
| Documentation coherence | Audited | 6 Markdown files, 38 local links, ID/table/fence/whitespace checks; active pointers and implementation checkboxes preserved | Reaudit affected documents after amendments |

## Planning evidence record

- Inspected baseline: `6128d312921cecca2a024bc8647d959126822f0b`, branch `main`, initially clean; 2026-09-08.
- Read target, masterplan, governance/rules, concepts, all feature outlines, relevant core implementation/test contracts, Makefile/scripts and research. No source, tests, dependencies or generated code changed.
- Official documentation and selected module manifests inspected; no driver installed, compatibility build or application tests executed.
- Product planning completed with G1–G4 still open. Exact dependency/toolchain compatibility is not asserted from a module fetch.
- Documentation structural audit: **passed**, 2026-09-08. A temporary documentation-only `planning-audit` recipe was added to the Make invocation with `--eval` while loading the canonical Makefile; no repository target or application test was added. Checked 6 Markdown files, 38 local links, balanced fences, table shapes, Markdown whitespace, 29 requirement mappings, 23 complete unit field sets, 73 unique unchecked scenarios, and 20 issue records. Active phase/target/completion and every existing master implementation checkbox were compared to HEAD and preserved. `git diff --check` passed. Existing intentional two-space Markdown line breaks were retained.
- Final confidence review: contract ownership, unit sequencing and scenario coverage are explicit. Corrected a dependency-cycle risk by separating G2/G3 admission decisions from the later runtime/measurement evidence their units produce. Full implementation readiness remains gated by G1–G4; no numerical confidence score substitutes for their evidence.
- Fresh local software validation: **not run**. Historical Phase 001 evidence remains historical. Hosted/manual/release evidence: **not run**.

## Implementation release gate

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
