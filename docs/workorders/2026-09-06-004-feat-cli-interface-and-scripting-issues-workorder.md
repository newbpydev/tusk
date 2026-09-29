---
feature-id: "004"
plan-source: docs/plans/2026-09-06-004-feat-cli-interface-and-scripting-plan.md
verification-plan: docs/verification-plans/2026-09-06-004-feat-cli-interface-and-scripting-verification-plan.md
status: Locally accepted - native and hosted release pending
evidence-scope: Local acceptance complete; PR 4 remediation locally verified; native and hosted release pending
---

# Feature 004 Issue Workorder

Companion to the [plan](../plans/2026-09-06-004-feat-cli-interface-and-scripting-plan.md) and [verification plan](../verification-plans/2026-09-06-004-feat-cli-interface-and-scripting-verification-plan.md). [MASTERPLAN.md](../../MASTERPLAN.md) controls execution.

Fixed in plan means the decision/coverage was corrected, not that implementation passes. Open execution gate means a specified implementation obligation still needs proof. Implementation was authorized on 2026-09-28; all seven units have local acceptance evidence; ISS-023 closes under the owner-delegated distribution policy. Owners below are project roles, not a claim that a particular human accepted an assignment.

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

All findings were raised during planning on 2026-09-09. V-IDs use the 004 prefix.

| ID | Source / owner lens | Severity | Status | Impact and decided next action | Retest / evidence |
| --- | --- | --- | --- | --- | --- |
| 004-ISS-001 | Outline; planning owner/coherence | P1 | Fixed in plan | False readiness and missing companions; supply requirements, units, failure contracts and linked pack | R25; U1–U7; document audit |
| 004-ISS-002 | Product R18; scope owner | P1 | Fixed in plan | Outline omitted history; include GetTaskHistory and wire shape in U3/U5 | R12–R13; V21,V74 |
| 004-ISS-003 | Product R16/R17; query owner | P2 | Fixed in plan | Range/velocity prose conflicts with current service; preserve day predicate and retained completions | R10,R12; V69,V72,V73 |
| 004-ISS-004 | MASTERPLAN/outline; architecture owner | P1 | Fixed in plan | TUI/completion would jump phases; explicit Phase 005/006 ownership, no stubs/default completion | R1,R24; V01,V85 |
| 004-ISS-005 | main.go; CLI owner | P1 | Fixed in plan | Scaffold accepts unknown commands; typed parse/domain split before data admission | R1–R2,R15; V01–V06 |
| 004-ISS-006 | service/options/path; composition owner | P1 | Fixed in plan | Configuration precedence and lazy resource owner absent; KTD2/KTD3 decide them | R3–R4,R16; V07–V13 |
| 004-ISS-007 | Service mutation port; API owner | P1 | Fixed in plan | Omitted/clear/latest patch semantics unspecified; use one UpdateTask with Base nil | R7–R8; V44–V51 |
| 004-ISS-008 | DeletePreview; consent owner | P1 | Fixed in plan | Force/recursion and stale preview unsafe if conflated; isolate deletion as U7 | R9,R20; V52–V63 |
| 004-ISS-009 | core omitempty; JSON owner | P1 | Fixed in plan | Direct marshaling loses fields/nulls/arrays; explicit DTO v1 converters | R13; V16–V22 |
| 004-ISS-010 | Process output; reliability owner | P1 | Fixed in plan | Writer/close failures after commit can invite duplicates; preserve committed outcome and never replay | R14–R16; V23–V27,V78–V80 |
| 004-ISS-011 | Durable recovery learning; error owner | P1 | Fixed in plan | Cause-first mapping hides transaction uncertainty; errors.As precedes errors.Is | R15; V25,V27,V80 |
| 004-ISS-012 | User text/terminal; security owner | P1 | Fixed in plan | Raw control/OSC/bidi text can manipulate terminal; sanitize only human display | R18–R20; V33–V35,V55,V62 |
| 004-ISS-013 | Lipgloss source; terminal owner | P2 | Fixed in plan | Global renderer/background queries violate isolation/startup; explicit invocation renderer/profile | R17–R19; V28–V39 |
| 004-ISS-014 | Makefile build; build owner | P1 | Fixed in plan | File-only build omits sibling signal/composition/tzdata code; package build + output override | R5,R23; V14,V15,V84 |
| 004-ISS-015 | Planned unit DAG; architecture owner | P2 | Fixed in plan | Consumers need output contract before implementation; preserve IDs, reorder U1→U5→U4→U2→U7→U3→U6 | R25; per-unit audit |
| 004-ISS-016 | Product performance protocol; evidence owner | P1 | Fixed in plan | Retain 100/1000 tasks and every sample; current owner-delegated distribution policy requires three runs and explicit tail/max guards | R22; V87–V89 |
| 004-ISS-017 | Headless coherence pass; docs owner | P2 | Fixed in plan | Product/registry/master pointers would still describe 004 as a stub; synchronize all affected triplets | R24–R25; link/status/checkbox audit |
| 004-ISS-018 | Headless feasibility pass; CLI owner | P1 | Fixed in plan | Cobra help callbacks can swallow writes; buffer help/version and route checked writer | R1,R14–R15; V01,V24 |
| 004-ISS-019 | Headless design pass; terminal owner | P2 | Fixed in plan | Width 1 cannot fit a wide cluster or deep indent; escaped fallback and bounded indent guarantee progress | R17–R18; V30–V34,V36 |
| 004-ISS-020 | Encoding/test scope; API owner | P2 | Fixed in plan | int64 history sequence could lose precision through float64 test decoding; integer-aware round-trip | R13; V21 |
| 004-ISS-021 | Chosen module sources; U1 build owner | P1 | Closed locally in U1 | Prove combined CLI graph/Go minimum/full-package targets before U1 closes | R5,R23; V14,V15,V84; module graph/build logs |
| 004-ISS-022 | Signal/terminal lifecycle; U1/U7/U6 owners | P1 | Closed locally (U6) | Prove bounded cancellation, joined prompt reader and actual EPIPE behavior; native console remains separate | R15,R20–R21; V57,V62,V78,V79,V86 |
| 004-ISS-023 | Product G3; U6 performance owner | P1 | Closed locally (U6) | Three complete runs meet fixed p90/tail guards without discarding samples | R22; V87–V89; raw samples, byte checks, manifest |
| 004-ISS-024 | Product G4; Feature 006 release owner | P2 | Open release gate | Run exact candidate natively and hosted; retained obligation blocks release, not CLI implementation | R23; V90–V91; native logs/job URLs/hashes |
| 004-ISS-025 | Feature 003 handoff; U6 integration owner | P1 | Closed locally (U6) | Prove actual process/disk workflow, unknown recovery, concurrent writers and consumer docs | R21,R24–R25; V76–V85; exact-SHA receipts |

## Open gate details

### 004-ISS-021: Combined dependency and executable proof

- **Phase found:** Planning. **Owner:** U1 build/compatibility. **Severity:** P1.
- **Evidence:** CLI libraries are absent from current go.mod; source manifests alone do not prove minimum-version selection, full executable startup or target compilation.
- **Expected:** Go 1.25 and accepted SQLite/libc retained; selected CLI graph builds as complete cmd/tusk package with timezone data and immutable version.
- **Action:** U1 adds pins, build/test targets and negative script tests; record actual red/green, combined graph and five target compile logs.
- **Closure:** V14,V15,V84 plus make validate. Revisit during U1 before its commit; blocks U1 completion, not planning.
- **Failure disposition:** No silent runtime/UI major-version change; make a concrete finding if selected graph fails.

### 004-ISS-022: Prompt cancellation and OS pipe behavior

- **Phase found:** Planning. **Owner:** U1 signal, U7 consent, U6 process verification. **Severity:** P1.
- **Evidence:** Current main has no context/stderr/signal contract. os/signal documents default broken-stdout behavior distinct from normal exit 1. Blocking terminal input needs its own cancellation seam.
- **Expected:** No hidden input on scripts; prompt error/cancel cannot delete; no leaked reader; confirmed mutations survive output errors.
- **Action:** Implement KTD5/KTD7 within process/terminal adapter ownership; direct fakes first, then actual Linux PTY/SIGPIPE/SIGINT/termination fixtures.
- **Closure:** V57,V62,V78,V79,V86 and canonical gates; Windows/macOS native proof remains ISS-024.
- **Blocking effect:** U1/U7 must close their local boundary cases; U6 must supply actual Linux acceptance. Revisit at each named unit. No manual waiver exists.

### 004-ISS-023: Reference latency and benchmark integrity

- **Phase found:** Planning. **Owner:** U6 performance maintainer. **Severity:** P1.
- **Evidence:** Feature 003 service samples exclude process startup and CLI output; no CLI benchmark runner exists. Existing product protocol requires empty/100/1000-task fixtures, five warmups and 100 measured launches.
- **Expected:** Every case in all three complete runs meets the current p90/tail/max policy, with valid fully consumed output. No result filtering.
- **Action:** Freeze host/toolchain/filesystem/binary hash/output fixture before measurement; collect raw samples and classify first-use/stress/slow pipe/lock observations separately.
- **Closure:** V87–V89 and runner negative tests. Revisit at U6; blocks local phase acceptance.
- **Failure disposition:** Fix a measured cause or request an explicit product-contract revision; a narrower dataset or p95 substitution is not closure. No claim of universal hardware/volume bound.

### 004-ISS-024: Native and hosted release acceptance

- **Phase found:** Planning. **Owner:** Feature 006 release maintainer. **Severity:** P2.
- **Evidence:** Linux inspection and cross-compilation cannot execute Windows/macOS console, signal/path/ACL and architecture behavior.
- **Expected:** Exact candidate native/hosted evidence and terminal lifecycle record before publication.
- **Action:** Import V90–V91 into Feature 006 planning; include Feature 002/003 carried native obligations, completion/man pages and release artifacts.
- **Closure:** Candidate SHA/hash, job URLs and target runtime records. Revisit when Phase 6 activates and before publication.
- **Blocking effect:** Release only; remains visibly pending after local CLI acceptance. This planned phase boundary is not an approved release waiver.

### 004-ISS-025: Real workflow and consumer handoff

- **Phase found:** Planning. **Owner:** U6 integration/documentation. **Severity:** P1.
- **Evidence:** Service port and disk recovery exist, but there is no CLI package or end-to-end executable workflow.
- **Expected:** Commands actually compose with the service and preserve results across independent processes; JSON/readback/script action parity complete.
- **Action:** Build once via Make in temporary output; execute full lifecycle, disk races, unknown outcome and failure cases; document exact grammar/config/DTOs and postcommit recovery.
- **Closure:** V76–V85, full gates, reviewed docs/cli.md and synchronized service/Feature 005/006 handoff.
- **Blocking effect:** Local phase acceptance; revisit U6. Existing service results cannot substitute.

## Review-lens sign-offs

Review is planning-only, performed sequentially in the main agent under the repository's tool mapping. No subagents or external peer were dispatched; no independent/cross-model corroboration is claimed.

| Lens | Status | Evidence / corrections | Runtime evidence still required |
| --- | --- | --- | --- |
| Architecture/dependency sequencing | Reviewed | Live ports, lazy factory boundary, stable-ID reordered DAG; ISS-004,006,014,015 | U1 combined graph, per-unit gates |
| Product/scope traceability | Reviewed | Existing product R1–R23/R27–R29; history restored, day/stats corrected, later phases explicit | Action parity V76,V85 |
| Coherence | Reviewed, corrections applied | Requirement groups, unit/scenario ownership, product pointers and linked triplets | Final document audit |
| Feasibility/reliability | Reviewed, corrections applied | Output/close ownership, help writer, unknown-cause precedence, SIGPIPE and input lifecycle | ISS-021,022,025 |
| API/agent contract | Reviewed | Exact JSON keys/nulls/arrays/order; opaque IDs, clear fields and int64 decoding | V16–V27,V40–V75 |
| Security/privacy | Reviewed | Local single-user scope; argv/stored controls → sanitizer; driver errors → allowlisted diagnostics; stale consent → authoritative service check | V27,V33–V35,V59–V63,V81 |
| Terminal design/accessibility | Reviewed, correction applied | Plain/TTY/empty/narrow states, width-1 fallback, labels without color and default-No prompt | V28–V39,V86 |
| Data integrity/concurrency | Reviewed | Reuse accepted service transactions, no split compound edit, no automatic replay, fresh owner recovery | V47,V59–V61,V79–V82 |
| Performance/resource behavior | Reviewed with execution gate | Product fixtures and maximum mandate retained; process/pipe timing and memory scope explicit | ISS-023 |
| Portability/operations | Reviewed with release gate | Complete-package builds, immutable version/tzdata, native evidence split | ISS-021,022,024 |
| Simplicity/maintainability | Reviewed | No config framework, generic DTO mapper, per-command repository or speculative shared TUI abstraction | Unit code reviews |
| Test strategy/evidence | Reviewed | Every unit has named red scenarios and Make gates; no software passes inferred from documents | All 91 scenarios pending |

Headless review state: five selected ce-doc-review lenses completed sequentially (coherence, feasibility, scope, security, terminal design), with the other required Ultrathink lenses covered in the same main-thread pass. Three draft-review corrections were applied: benchmark fidelity, checked help/version writes, and width-1 fallback. No unresolved proposed product change or decision remains. Twenty planning findings are fixed in the documents; five execution/release gates remain. These are implementation obligations, not evidence of current defects in nonexistent CLI code.

## Planning audit record

- Baseline: initially clean main at 85bf1171ede438c8e52f7d19298fef51c9d3640d, Linux x86_64, inspected 2026-09-09.
- Inputs: target outline, user/checked-in AGENTS, masterplan, product and Feature 003 packs, service handoff, core/ports/service/storage/main/Makefile, applicable cursor rules, durable transaction outcome learning.
- External evidence: selected Cobra/Lipgloss/terminal manifests and source, Go signal/timezone docs; links and their limited claims are in the plan. No dependency installation/runtime test performed.
- Artifacts: 25 requirements; seven units U1→U5→U4→U2→U7→U3→U6; 91 unchecked scenarios (89 local, two release handoffs); 25 findings.
- Confidence recheck: command intent, output/resource outcomes, direct service mapping, unit boundaries and evidence ownership are explicit. No planning blocker remains; actual compatibility/latency/OS behavior requires implementation evidence.
- Document audit: PASS on 2026-09-09 via Make temporary planning-audit-004 target: eight Markdown files, 99 local links/anchors; 25 requirements, seven ordered units, 91 unchecked scenarios and 25 findings. Product requirement/grammar text and all 73 product scenario check states are unchanged; only Phase 4 planning boxes advanced. git diff --check passes; documentation-only changes and no staged files.
- Application tests/make validate/build/benchmarks: not run during planning. No implementation, release, product scenario or prior-phase acceptance checkbox is newly completed by this pack.
- Commit/push/PR/merge/publication: not performed in this planning pass.

## Implementation release gate

Leave every item unchecked during planning.

- [x] U1, U5, U4, U2, U7, U3, U6 implemented with per-unit red/green, make validate and separate commits.
- [x] Local V01–V89 executed with exact-revision receipts.
- [x] ISS-021,022,023,025 closed with local evidence.
- [x] CLI/main/harness coverage at least 95%; generated output and prior service/storage behavior unchanged.
- [x] Reference latency gates pass without discarding samples.
- [x] Real Linux terminal/process behavior recorded.
- [x] Feature 005/006 handoff and all planning/checklist artifacts synchronized.
- [ ] ISS-024 / V90–V91 exact candidate native/hosted release acceptance complete before publication.
- [x] No unresolved local blocker; any observed acceptance change has explicit decision-owner approval.

## Execution checkpoint — 2026-09-28

U1 headless evidence is in the [execution receipt](../verification-evidence/004/README.md).
ISS-021 has passing dependency/minimum-Go/cross-build evidence, but U1 acceptance
and commit remain pending. ISS-022 still needs real terminal/process evidence in
its declared units. No later unit or release gate is complete.

### 004-ISS-026: Required Kitty verification cannot reach the desktop

- **Owner:** U1 execution; **severity:** P1; **status:** Open execution blocker.
- **Requirement:** User instruction on 2026-09-28, now recorded in AGENTS.md:
  always verify using Kitty and inspect the current behavior against the plan.
- **Observed:** Kitty is installed. Wayland reports `Failed to connect to display`;
  explicit X11 reports `Failed to open display :0`; both exit 1 before checks run.
- **Disposition:** Headless tests remain useful partial evidence, but do not close
  Kitty acceptance. No existing user terminal was changed. Resume from U1 in a
  session with desktop access, run the retained verification script in an owned
  Kitty window, inspect help/version/error behavior, retain the terminal result,
  review U1, synchronize this triplet/masterplan, and commit before advancing U5.
- **Git state:** Initial staging succeeded, but restoring those owned paths to the
  unstaged state failed: `Unable to create .git/index.lock: Read-only file system`.
  The implementation remains staged, with later documentation updates unstaged.
  No commit was attempted because Kitty acceptance is still pending. Preserve
  both index and working tree; restore normal Git write access before the boundary
  commit. No workaround outside the configured permission boundary was attempted.
- **Root cause confirmed:** [Permission diagnosis](../verification-evidence/004/session-permissions.json)
  records Kitty 0.49.1, existing Wayland/X11 sockets and readable Xauthority, with
  both socket connections denied by EPERM. The Git mount is explicitly read-only.
  Restore session permission for desktop connections and Git writes; reinstalling
  Kitty is unnecessary. The agent cannot change the enclosing session sandbox.

### U1 acceptance after permission restoration

ISS-026 is closed: restored session permissions allowed an owned Kitty window,
canonical gates and actual help/version/error inspection. Git is writable again.
ISS-021 is closed locally by the retained Go 1.25/full graph/five-target evidence.
ISS-022 still needs the later consent and real mutation/pipe cases. See [U1 acceptance](../verification-evidence/004/u1-accepted.json).
The unit was reviewed sequentially for architecture, contracts, dependency
compatibility, reliability and portability; no blocking finding remains. No
independent peer review or release acceptance is claimed.

### U5 local acceptance — 2026-09-28

Explicit JSON DTOs preserve complete task and query data, nulls and arrays, UTC times and exact history sequence numbers. Output failures retain known committed state; unknown outcomes and private error redaction remain intact.

Canonical `make validate` passed in an owned Kitty window; current CLI output
was inspected there. [Receipt](../verification-evidence/004/u5-accepted.json)
retains red/green logs and source hashes. Scenarios 16-27 are locally
verified; native release acceptance remains separate. Advance to U4
only after this unit commit.

### U4 local acceptance — 2026-09-28

Human output escapes terminal controls and bidi directives, preserves complete IDs, wraps by grapheme cell width, and uses invocation-owned styling without background probes. Actual Kitty inspection found and fixed header alignment; the six-cell terminal priority header is PRIO, while TSV remains PRIORITY.

Canonical `make validate` passed in an owned Kitty window; current CLI output
was inspected there. [Receipt](../verification-evidence/004/u4-accepted.json)
retains red/green logs and source hashes. Scenarios 28-39 are locally
verified; native release acceptance remains separate. Advance to U2
only after this unit commit.

### U2 local acceptance — 2026-09-28

Add, edit and done map exact argument intent to one accepted service call. Disk integration proves subtree lifecycle, atomic combined patches, no-op history, parent policy, and supported service dates (+1d/+1w/+1m). The built CLI was exercised in Kitty on an isolated database. Shared-code reuse, quality and efficiency review ran sequentially with no behavior-preserving change warranted.

Canonical `make validate` passed in an owned Kitty window; current CLI output
was inspected there. [Receipt](../verification-evidence/004/u2-accepted.json)
retains red/green logs and source hashes. Scenarios 40-51 are locally
verified; native release acceptance remains separate. Advance to U7
only after this unit commit.

### U7 local acceptance — 2026-09-28

Deletion requires independent recursion and force intent, defaults to no in an eligible terminal, and passes unchanged preview consent to the service. Second-owner add/remove/move/metadata races reject stale consent. Unix input uses bounded polling; Windows cancellation joins its pinned reader, including the no-pending-I/O race. Actual Kitty decline, acceptance and recursion refusal passed.

Canonical `make validate` passed in an owned Kitty window; current CLI output
was inspected there. [Receipt](../verification-evidence/004/u7-accepted.json)
retains red/green logs and source hashes. Scenarios 52-63 are locally
verified; native release acceptance remains separate. Advance to U3
only after this unit commit.

### U3 local acceptance — 2026-09-28

List, tree, stats and metadata history consume authoritative service read models without resorting or recomputing. Tests cover exact filters, empty/missing results, retained completions and failure suppression. Actual Kitty query inspection improved long history output to labeled records when aligned columns cannot fit.

Canonical `make validate` passed in an owned Kitty window; current CLI output
was inspected there. [Receipt](../verification-evidence/004/u3-accepted.json)
retains red/green logs and source hashes. Scenarios 64-75 are locally
verified; native release acceptance remains separate. Advance to U6
only after this unit commit.


## U6 measured performance remediation (2026-09-28)

The first full reference runs failed the 1,000-task list/tree bounds on both
Go 1.27.1 and minimum Go 1.25.0; all raw samples are retained. Profiling attributes
substantial time to task decoding, timestamp parsing/formatting and allocation.
U6 therefore includes a narrow storage codec optimization, without schema or
service-contract changes. A failing zero-allocation timestamp test precedes the
change; exhaustive single-byte canonical-format parity and existing corrupt-row
tests guard disk compatibility. This is an explicit performance remediation of
ISS-023. The latency gate remains pending until a complete reference run passes;
no fixture reduction, percentile substitution or discarded outlier is allowed.

## U6 execution checkpoint — 2026-09-28

The [U6 receipt](../verification-evidence/004/u6-checkpoint.json) and
[evidence index](../verification-evidence/004/README.md) bind source hashes to
canonical logs, actual process/PTY tests, Kitty screenshots and complete raw reports.
`make validate build check-generated` passes, including race and coverage gates;
minimum Go full tests and five CGO-free executable/test builds pass. V01–V86 and
V88–V89 are locally verified. V90–V91 remain Feature 006 release obligations.

ISS-022 and ISS-025 are closed locally: actual executable lifecycle, unsafe-path
refusal, concurrent writers, broken stdout/stderr, cancellation, unknown outcomes
and hard-kill readback pass. Actual Linux PTY tests cover yes/no/EOF/SIGINT/SIGTERM
and reaping. Visible Kitty output was inspected at 40/80/120 columns with Unicode
and color-disabled modes. ISS-026 stays closed; Kitty is required by AGENTS.md.

Review corrections enforce signed decimal progress and reject incomplete benchmark
fixtures. Narrow timestamp and output-copy allocation fixes preserve storage and
wire contracts. V87 is **failed**, not waived: final 1,000-task list/tree JSON maxima
are 30.118/32.832 ms, each violating 15 ms in all 100 samples. Help/version, empty
and 100-task cases pass in that run. Earlier failed runs and all samples are retained.
First-use, 10k, 1 MiB, slow-output and held-writer observations are separate V88
records. V89 rejects incomplete fixtures, wrong counts, failed children and bound
violations. No dataset reduction, percentile substitution or outlier removal occurred.

ISS-023 remains open (P1); U6 and Phase 4 are not accepted. Further measured
optimization or an explicit owner decision is required. The completed review and
resolution record retain finding 1 as open and findings 2/3 as fixed. No contract
revision, Phase 5 advancement or publication is authorized by this checkpoint.

[CLI documentation](../cli.md) supplies grammar, JSON fields, configuration, consent
and recovery. Service and product documents carry the same handoff: Feature 005
owns TUI registration, draft/refresh/consent behavior; Feature 006 owns native/hosted
runtime, completion/man pages and release proof (ISS-024 / V90–V91 still open).

### Owner-directed latency remediation — 2026-09-28

The owner explicitly chose to retain 15 ms and continue broader storage/query
optimization. U6/ISS-023 remains the active target. Measured decoding, allocation,
snapshot and serialization costs may be optimized across storage/service/CLI;
public ports, detached results, corruption checks, wire data, transaction outcomes
and every-sample bounds remain unchanged. Each implementation change needs an
observed performance/behavior regression test and the complete canonical gate.
The checkpoint measurements above precede this remediation and do not identify
later working-tree edits. Rebuild and retain a complete new reference run before
claiming V87 or U6 acceptance; no schema change or later-phase work is authorized.

### Remediation implementation notes

Measured fixes now cover canonical tag decoding with exhaustive corruption/error
parity, batch decoding directly into values, reuse of detached service snapshots,
avoidance of unfiltered storage clones, and allocation-free recognition of already
ordered snapshots in the core sorter. The ordering contract and detached results
are unchanged. Per-open schema inspection reuses only the expected embedded-DDL
catalog; it still reads and verifies the actual catalog and ledger on every check.
The initial fixes changed no schema, SQL query or generated sqlc output; the subsequent ListAll optimization below explicitly adds a generated read query.

Task JSON uses the explicit DTO fields and appends into one owned buffer, with
standard encoding for escaped/non-ASCII strings and byte-for-byte parity tests
against encoding/json. Nullable fields, UTC timestamps, empty arrays, errors and
one final LF remain unchanged. Safe human text can bypass copying; control/bidi
escaping and invalid UTF-8 replacement remain covered.

Historical scheduling trial: a diagnostic measured approximately 11.0 ms versus
9.1 ms for a serial query with six versus one Go execution processors. That trial
set a single-processor default in the standalone executable while respecting
explicit GOMAXPROCS. The override was removed before U6 local acceptance; both
the executable and embedded cli.Run preserve Go's runtime scheduling policy.
The diagnostic remains historical evidence and does not describe the accepted
binary or require Feature 005 to inherit a processor-count override.

### Unfiltered SQL optimization

The current profile attributes about one third of CPU time to ListCandidates.
U6 adds a sqlc-generated `ListAll` query (`SELECT * FROM tasks`) only when every
filter is empty, eliminating unnecessary optional-predicate evaluation/binding.
The filtered path, strict row decoder, detached results and canonical ordering
remain authoritative. A failing missing-query test precedes `make generate`;
empty/full-row parity and all existing filter/corruption tests guard this change.
This is an explicit query/generated-output amendment under the owner's broader
optimization instruction, with no schema or migration change. `check-generated`
and the full gate must pass again before acceptance.

### Generated row allocation and benchmark consumer correction

The owner-directed storage optimization enables sqlc's
`emit_result_struct_pointers` and regenerates query results. This removes copying
large row structs while result slices grow; a 100-row allocation regression
failed at 101,051 bytes before the change and now passes a 64 KiB bound. Complete
row parity, strict decoding, generated consistency and fault propagation remain
required. Ports, schema and migration contracts are unchanged.

A separate diagnostic retained all samples and measured child CPU alongside wall
time. In the baseline pipe consumer, parent garbage collection occurred during
each measured 1,000-task JSON invocation; the largest wall sample was 30.182 ms
while child CPU was 12.53 ms. Other samples had genuinely elevated child CPU.
The harness now allocates 1 MiB stdout/4 KiB stderr capacity and collects prior
validation garbage before launching the timed child. Buffers can still grow;
child GC is unchanged. Launch, process exit and full pipe drainage stay timed,
and all 100 consecutive samples must still meet the original strict limit.
The manifest records this preparation. Earlier failed reports remain evidence;
this correction alone does not establish latency acceptance.

### Stable ordering movement optimization

The current single-processor profile attributes 30% of query CPU to reflective
stable sorting and its repeated pointer-bearing Task moves. SortTasks now stably
sorts integer positions, then applies permutation cycles in place. Each task is
moved at most once into its final position, plus one saved value per cycle. The
already-ordered path still allocates nothing. A randomized full-value parity
test includes duplicate IDs/keys to protect stability, sizes 0–1,000, and a
one-buffer allocation bound (observed Red: three allocations; Green: one).
Canonical priority/date/created/ID semantics are unchanged.

## Broader optimization checkpoint — 2026-09-28

The owner retained the **15 ms** requirement and authorized broader storage/query
optimization. That decision is settled; no bound relaxation is pending.

[Current source hashes and results](../verification-evidence/004/u6-broader-checkpoint.json)
identify this uncommitted checkpoint. `make validate build check-generated`
passes in Kitty (main 95.8%, CLI 96.8%, core 98.3%, service 95.6%, storage 97.8%,
benchmark 95.5%). Go 1.25 full tests and five CGO-free executable/test cross-builds
also pass. Fresh visible Kitty output confirms decimal progress, syntax exit 2,
and persisted readback. Cross-builds remain distinct from native runtime proof.

The post-sort balanced reference run retains every sample. Its 1,000-task
list/tree medians are approximately 9.6–10.8 ms, but maxima remain 19.306 ms
(human list), 17.752 ms (JSON list), 12.594 ms (human tree), and 18.818 ms
(JSON tree). Some smaller-fixture cases also have isolated failures. Minimum-Go,
CPU-affinity and temporary performance-profile runs fail too; the last of these
has a 26.164 ms JSON-list maximum. The host returned to balanced mode afterward.
None of these reports is accepted or substituted for a passing reference run.

Separate diagnostics retain wall time, child user/system CPU, parent GC counts,
and experimental settings. Higher GC thresholds and processor counts did not
resolve the failures, so no GC policy change was applied. GC tracing recorded no
child collections in the traced samples. Temporary stage instrumentation shows
small output-write/signal-stop costs and variable storage/query costs; it does
not establish a sole cause or a latency guarantee. Diagnostic fixtures and
instrumentation are explicitly separate from acceptance evidence.

**U6, V87, ISS-023 and Phase 4 remain open.** The original completed review receipt
predates the broader changes; a fresh review is required before committing this
unit. Continue from the current working tree, retain the exact fixture and
sample rules, resolve the latency failures, then review, synchronize acceptance
and commit U6. Phase 5 remains blocked. No push, PR or publication occurred.

### Reader connection reuse — active U6 continuation

Open now retains its physically read-only inspection connection as the reader
pool after compatibility and writer migration checks. This reduces physical
connections from three to two while retaining separate writer/reader pools,
query-only pragmas, four-reader/one-writer bounds, strict catalog checks and
replacement-connection configuration. There is no port or schema change.

The allocation/lifetime test first failed with three opened/closed handles.
The full suite then caught stale journal-mode state after first-use migration;
Open now reads the schema version to refresh the retained pager before exposing
the pool. The existing cross-process WAL snapshot test passes again. Fault tests
cover refresh failure and closing both pools when the retained reader close
fails. First-use and existing-file tests verify read-only enforcement and exact
handle closure. Latency acceptance remains pending a new reference run.

### Generated expected catalog — U6 storage amendment

Fresh-process CPU profiles attribute about 18% of sampled CPU to recreating the
expected schema. U6 now embeds a generated catalog for every migration prefix,
produced by the pinned SQLite runtime in private memory. Exact SHA-256 hashes of
the SQL bytes select it; unknown/changed inventories fall back to live private
memory evaluation. No user database contents are cached. The actual identity,
ledger, and full catalog are still read and compared in their existing snapshots.

`make generate` regenerates the catalog atomically; `make check-generated` and
normal full tests independently compile every prefix and compare exact bytes.
This amends Feature 002's evaluation timing for immutable expected data only;
SQLite remains the schema compiler and there is no handwritten DDL description,
new daemon, package-level mutable cache, schema change, or relaxed drift check.
A failing allocation test observed 133 allocations; generated lookup uses 18.
Mismatch, stale-checksum/changed-SQL, cancellation, and dynamic fallback tests
pass. The first 200-allocation test passed and is baseline evidence, not Red.

The reader-connection reuse experiment above was removed: a complete reference
run showed no meaningful median improvement despite its extra initialization
logic. Its failure, fix, full gate, raw samples and patch remain diagnostic
evidence. Production Open retains separate inspection, writer and reader handles.
The 15 ms gate remains unchanged and pending the next complete run.

### Verification execution clarification — 2026-09-28

The owner clarified that canonical checks and latency measurements should run in
Codex's default Bash. Kitty remains required for the visible terminal scenarios
in this verification pack, with retained output and interaction evidence.
AGENTS.md now states that distinction. Earlier Kitty command logs remain valid
historical records; new latency runs record Bash as the launching environment.
The strict 15 ms query bound and all retained-sample rules remain unchanged.

The generated catalog passed `make validate build check-generated` before the
subsequent generation-order fix; that fix passed generation and all 34 script
checks. A full current-source gate is still required. PGO did not materially
improve the reference medians and is not enabled in production builds. An
isolated JSON row-transfer experiment was over twice as slow as typed column
reads and was discarded. The tree builder now allocates its nodes in one batch;
a red test observed 1,026 allocations for 1,000 roots against a 100-allocation
bound, followed by a passing full short suite. Existing detached-copy and graph
validation tests remain in force. U6 acceptance is still pending.

### Historical U6 Bash checkpoint — 2026-09-28

`make validate build check-generated` passed in Codex Bash, including race,
coverage and 34 script checks. Go 1.25 `make test-unit build-cli` passed for all
five compilation targets. Fresh actual Kitty inspection confirms tree depths,
10% decimal progress, child completion and parent rollup with auto-completion
disabled, full IDs, aligned columns and statistics. Evidence surfaces remain
separate under the owner's clarified policy.

The balanced Bash reference run has 1,000-task list/tree medians of 9.16–10.66 ms
but maxima of 16.59–18.81 ms. A small human-history case also has a 22.14 ms
outlier. The temporary performance-profile run still fails (including one
5.32 ms help sample); the profile returned to balanced. Every raw sample is
retained. Neither Kitty nor output-pipe capacity is established as the cause.
No threshold, sample count or fixture was weakened. U6/V87/ISS-023 remain open;
there is no acceptance, unit commit or Phase 5 advancement. Fresh code review
is still required once performance changes settle.

[Current source hashes and receipts](../verification-evidence/004/u6-bash-checkpoint.json)
include the Bash gates, all new reference reports, diagnostic experiments,
red/green tests and visible terminal evidence. The expected-catalog allocation
regression and batched-node regression are green; discarded experiments are
retained as evidence only, not enabled in production.

## U6 local acceptance — 2026-09-28

U6 and Feature 004 are locally accepted under the owner-delegated distribution
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

## Post-acceptance branch simplification review — 2026-09-28

The owner requested `ce-simplify-code` for the branch. Three review lenses ran
inline per the project tool mapping. A shared human-tree output buffer halved
allocation in the deep-tree fixture and passed functional/race/Kitty checks.
The trial passed 83/84 latency case-runs; one unchanged JSON-tree case had p90
15.125 ms. A pre-change source control passed all 84. Causation is unconfirmed,
so the trial was reverted under the skill's verification rule. No production or
test changes remain, and the original U6 acceptance source is preserved.
The [evaluation and retained evidence](../verification-evidence/004/simplify-review.md)
record the rejected patch, all samples and restored-source validation.
Feature 005 planning remains next; native/hosted release gates remain deferred.

## Final branch review follow-up — 2026-09-28

User-authorized P0–P2 review completed with two documentation fixes: the CLI
guide now preserves Go runtime scheduling defaults, and directs automated gates
and timing to Codex Bash while retaining Kitty for required visible scenarios.
The superseded single-processor trial is explicitly historical. The follow-up
diff review is clean; `make validate build check-generated` passes and all 158
accepted source hashes are unchanged. No fresh latency or Kitty result is claimed.
See [review receipt](../verification-evidence/004/review-final.md) for coverage,
peer availability and all requirement dispositions. Feature 005 remains planning
only; Feature 006 native/hosted release proof remains deferred.

### PR #4 review remediation — 2026-09-29 (locally verified)

The explicit babysit invocation authorizes this Feature 004 follow-up before
Feature 005 planning resumes. Scope: terminal escaping and capability detection,
trusted configuration/deletion hints, second-signal termination during stalled
cleanup, colocated filter semantics, reader ownership documentation, ordering
regressions, module metadata and benchmark tooling. EOF still declines deletion;
transaction uncertainty retains precedence. Native console proof remains V90.

Historical timing diagnostics that combined child settings with parent GC and
preallocation remain unchanged artifacts; they cannot isolate child-only effects.
New child diagnostic modes isolate those settings. New reference report fields
use snake_case; historical reports retain their original field names.

Red/green regressions, `make validate build check-generated check-modules`,
Go 1.25 full tests/five-target builds and owned Kitty inspection pass. The isolated
reference matrix passes all 84 case-runs with 8,400 samples retained (worst query
p90 12.495 ms). An earlier run overlapped final cross-builds and failed six cases;
its complete samples remain diagnostic evidence. See
[review dispositions and receipts](../verification-evidence/004/review-pr4-r1.md).
Hosted feedback settlement remains separate from these local results.
