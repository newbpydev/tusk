---
feature-id: "004"
plan-source: docs/plans/2026-09-06-004-feat-cli-interface-and-scripting-plan.md
verification-plan: docs/verification-plans/2026-09-06-004-feat-cli-interface-and-scripting-verification-plan.md
status: Planning reviewed - execution gates open
evidence-scope: Planning findings only
---

# Feature 004 Issue Workorder

Companion to the [plan](../plans/2026-09-06-004-feat-cli-interface-and-scripting-plan.md) and [verification plan](../verification-plans/2026-09-06-004-feat-cli-interface-and-scripting-verification-plan.md). [MASTERPLAN.md](../../MASTERPLAN.md) controls execution.

Fixed in plan means the decision/coverage was corrected, not that implementation passes. Open execution gate means a specified implementation obligation still needs proof. This pack has no unresolved product decision; it is ready to be implemented only after authorization. Owners below are project roles, not a claim that a particular human accepted an assignment.

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
| 004-ISS-016 | Product performance protocol; evidence owner | P1 | Fixed in plan | A smaller fixture/percentile-only result would weaken mandate; include 100/1000 tasks, five warmups, every sample | R22; V87–V89 |
| 004-ISS-017 | Headless coherence pass; docs owner | P2 | Fixed in plan | Product/registry/master pointers would still describe 004 as a stub; synchronize all affected triplets | R24–R25; link/status/checkbox audit |
| 004-ISS-018 | Headless feasibility pass; CLI owner | P1 | Fixed in plan | Cobra help callbacks can swallow writes; buffer help/version and route checked writer | R1,R14–R15; V01,V24 |
| 004-ISS-019 | Headless design pass; terminal owner | P2 | Fixed in plan | Width 1 cannot fit a wide cluster or deep indent; escaped fallback and bounded indent guarantee progress | R17–R18; V30–V34,V36 |
| 004-ISS-020 | Encoding/test scope; API owner | P2 | Fixed in plan | int64 history sequence could lose precision through float64 test decoding; integer-aware round-trip | R13; V21 |
| 004-ISS-021 | Chosen module sources; U1 build owner | P1 | Open execution gate | Prove combined CLI graph/Go minimum/full-package targets before U1 closes | R5,R23; V14,V15,V84; module graph/build logs |
| 004-ISS-022 | Signal/terminal lifecycle; U1/U7/U6 owners | P1 | Open execution gate | Prove bounded cancellation, joined prompt reader and actual EPIPE behavior; native console remains separate | R15,R20–R21; V57,V62,V78,V79,V86 |
| 004-ISS-023 | Product G3; U6 performance owner | P1 | Open execution gate | Freeze reference manifest, implement runner and meet every-sample bounds without trimming | R22; V87–V89; raw samples, byte checks, manifest |
| 004-ISS-024 | Product G4; Feature 006 release owner | P2 | Open release gate | Run exact candidate natively and hosted; retained obligation blocks release, not CLI implementation | R23; V90–V91; native logs/job URLs/hashes |
| 004-ISS-025 | Feature 003 handoff; U6 integration owner | P1 | Open execution gate | Prove actual process/disk workflow, unknown recovery, concurrent writers and consumer docs | R21,R24–R25; V76–V85; exact-SHA receipts |

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
- **Expected:** Every normal help/version sample <5 ms; every query/format sample <15 ms, with valid fully consumed output. No result filtering.
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

- [ ] U1, U5, U4, U2, U7, U3, U6 implemented with per-unit red/green, make validate and separate commits.
- [ ] Local V01–V89 executed with exact-revision receipts.
- [ ] ISS-021,022,023,025 closed with local evidence.
- [ ] CLI/main/harness coverage at least 95%; generated output and prior service/storage behavior unchanged.
- [ ] Reference latency gates pass without discarding samples.
- [ ] Real Linux terminal/process behavior recorded.
- [ ] Feature 005/006 handoff and all planning/checklist artifacts synchronized.
- [ ] ISS-024 / V90–V91 exact candidate native/hosted release acceptance complete before publication.
- [ ] No unresolved local blocker; any observed acceptance change has explicit decision-owner approval.
