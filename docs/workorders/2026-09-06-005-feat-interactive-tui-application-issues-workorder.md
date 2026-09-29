---
feature-id: "005"
plan-source: ../plans/2026-09-06-005-feat-interactive-tui-application-plan.md
verification-plan: ../verification-plans/2026-09-06-005-feat-interactive-tui-application-verification-plan.md
status: Open - execution gates pending
evidence-scope: Planning findings only
---

# Feature 005 Issue Workorder

The [plan](../plans/2026-09-06-005-feat-interactive-tui-application-plan.md) and
[verification plan](../verification-plans/2026-09-06-005-feat-interactive-tui-application-verification-plan.md) are one
planning pack. Work only in the sequence authorized by [MASTERPLAN.md](../../MASTERPLAN.md).

**Planning result:** 20 source/plan defects have explicit corrections in the
pack; five execution/release gates remain open or deferred. “Fixed in plan”
means the decision/verification obligation was corrected. It does not mean code
was fixed or any test passed. There is no unresolved product decision preventing
U1; U1 may start only when implementation is authorized.

## Planning issue register

**U2 owner direction, 2026-09-29:** the user selected spacious title/metadata
rows after the live Kitty design prototype and requested removal of modal/list
shadow artifacts. Carry uniform surfaces, whole-backdrop dimming and explicit
SGR reset handling into production. Prototype fixes and Bash cell-color checks
are complete; production layout and V39 Kitty evidence now pass. This
refines 005-ISS-018 within the accepted sequence; no new density setting or
scope expansion is authorized by the prototype comparison keys.

| ID | Source | Owner/lens | Severity | Status | Gap and correction | Affected scope / retest |
| --- | --- | --- | --- | --- | --- | --- |
| 005-ISS-001 | Planning/source review | Coherence / planning owner | P1 | Fixed in plan; runtime unverified | Outline falsely labeled ready; no companion artifacts or executable units. Replace outline with complete linked pack and preserve planning-only status. | R28; all units; Plan Goal Capsule, 28 requirements, eight units and verification coverage audit. |
| 005-ISS-002 | Planning/source review | Product scope / TUI owner | P1 | Fixed in plan; runtime unverified | Feature outline omitted product groups, timeline, debounce and 2-second refresh. Carry product R18/R24–R26 and KTD14 into feature contracts without adding a new service. | R8,R10,R11,R14,R28; U3/U4; V40–V67; product U16–U20 crosswalk. |
| 005-ISS-003 | Planning/source review | Compatibility / U1 owner | P1 | Fixed in plan; runtime unverified | Selected Bubble Tea needs newer ansi; newest checked Glamour changes Lipgloss pin. Choose v1.3.10/v0.21.0/v1.1.0/Glamour v0.9.1 and ansi v0.10.1; retain runtime pins. | R4,R25; U1/U6; Versioned module source reviewed; V01–V03,V105–V106 remain unexecuted. |
| 005-ISS-004 | Planning/source review | Purity / U1 owner | P1 | Fixed in plan; runtime unverified | Bubbles textarea View mutates a pointer-backed wrap cache; value receiver is not proof of purity. Prepare complete frame during constructor/Update; root View returns immutable string. | R5; U1/U2; Upstream textarea source; V05,V06,V37 include cold and populated deep-state checks. |
| 005-ISS-005 | Planning/source review | CLI contracts / U7 owner | P1 | Fixed in plan; runtime unverified | No tui registration; naive launch can open DB on help, pipes or syntax errors. Add optional injected runner after grammar/config/TTY admission; preserve bare help. | R1,R2; U7; Current root/config/composition read; V13–V17. |
| 005-ISS-006 | Planning/source review | Lifecycle / U7 owner | P1 | Fixed in plan; runtime unverified | Bubble Tea shutdown does not join all in-flight command goroutines. Close admission, cancel and drain command work before owner close; record outcome before reply delivery. | R3,R6,R22; U7; Versioned tea.go handleCommands/shutdown; V18–V25,V103. |
| 005-ISS-007 | Planning/source review | Async correctness / U3 owner | P1 | Fixed in plan; runtime unverified | Late data can replace new form/selection; timer invalidation can starve slow reads. Single service operation, coalesced flags, distinct relevance tokens, immutable captures. | R8,R24; U3/U4/U5; V07,V46,V50–V54,V61,V64. |
| 005-ISS-008 | Planning/source review | Hierarchy / U3 owner | P1 | Fixed in plan; runtime unverified | Filtering before tree construction loses ancestors or manufactures roots. Project complete tree using core predicates and inclusive/exclusive due-day bounds. | R10,R11; U3; Current GetTaskTree/core filtering read; V40–V49. |
| 005-ISS-009 | Planning/source review | Service contract / U4 owner | P2 | Fixed in plan; runtime unverified | Tree and history methods are separate snapshots; apparent joined freshness is unsupported. Give history separate token/loading/stale state; never claim atomic metadata/history. | R13,R14,R24; U4; Service queries contract; V63–V66. |
| 005-ISS-010 | Planning/source review | Terminal security / U2/U4 owners | P1 | Fixed in plan; runtime unverified | Stored text, Markdown links and renderer options can inject terminal output or hidden I/O. Share scalar sanitization, explicit multiline policy, fixed renderer style/profile and final escape filtering. | R15; U2/U4; Current CLI sanitizer and Glamour source; V31–V36,V59–V60. |
| 005-ISS-011 | Planning/source review | Data preservation / U5 owner | P1 | Fixed in plan; runtime unverified | Widget sanitization/line limits can silently rewrite original notes during metadata edits. Keep raw values and dirty fields; require lossless round-trip or explicit read-only replacement path; reject paste atomically. | R23; U5; Bubbles source; V33,V58,V70,V84–V85. |
| 005-ISS-012 | Planning/source review | Parity / U5 owner | P1 | Fixed in plan; runtime unverified | Forms could omit clear intent, misuse due display values or recompute business rules. Map only dirty fields to existing Base/clear/patch APIs; service owns graph/status/progress. | R16–R18; U5; ports.TaskService and service.md; V68–V80. |
| 005-ISS-013 | Planning/source review | Reliability / U5 owner | P1 | Fixed in plan; runtime unverified | Duplicate submit or saved-refresh-failed state can repeat an acknowledged mutation. One admission before waiting for read drain; known success closes draft, stale read blocks writes. | R19; U5; V81–V83,V20. |
| 005-ISS-014 | Planning/source review | Destructive UX / U8 owner | P1 | Fixed in plan; runtime unverified | One-key delete and changing subtree can broaden consent. Freeze preview, default Cancel, explicit recursive checkbox, Force=false; reset consent after conflict. | R20; U8; DeletePreview/Expected live contract; V87–V92. |
| 005-ISS-015 | Planning/source review | Transaction integrity / U8 owner | P1 | Fixed in plan; runtime unverified | Unknown outcome matches retryable causes; uncertain create may expose no ID. Check typed outcome first; retire/reopen same owner path, quarantine draft, read back and never guess/replay. | R21; U7/U8; Durable outcome-redaction learning; V21–V22,V93–V99. |
| 005-ISS-016 | Planning/source review | Process / U7 owner | P1 | Fixed in plan; runtime unverified | Competing signal handlers or lost result messages obscure committed state/cleanup failure. Keep process signal owner; raw-key cancellation is explicit; outcome receipt survives Program cancellation. | R22; U7/U6; Current processContext/CLI diagnostics read; V20,V23–V25,V98,V103. |
| 005-ISS-017 | Planning/source review | Reliability / U7 owner | P2 | Fixed in plan; runtime unverified | A cancellation deadline is not a guarantee that a service goroutine has ended. 10-second cooperative operation deadline; drain before close, report slow cleanup and retain escalation limitation. | R3,R6,R22; U7; V19,V25,V102–V103. |
| 005-ISS-018 | Planning/source review | Interaction design / U2 owner | P2 | Fixed in plan; runtime unverified | Centering with lipgloss.Place alone does not compose an overlay; undersize/modal behavior unspecified. Define cell compositor, exact panel/row bounds, scrolling/focus and draft-preserving small screen. | R9,R12; U2; V27–V39. |
| 005-ISS-019 | Planning/source review | Evidence / U6 owner | P1 | Fixed in plan; runtime unverified | View-only timing, headless checks or cross-builds can be mistaken for user-visible/native acceptance. Separate frame preparation, CLI process latency, PTY, owned Kitty and native/hosted proof. | R25,R26; U6; V103–V108,V111–V112; no performance or terminal claim during planning. |
| 005-ISS-020 | Planning/source review | Governance / docs owner | P2 | Fixed in plan; runtime unverified | Registry/product pointers describe delivered CLI as unstarted and Feature 005 as outline. Synchronize Feature 005 pack, product triplet, registry and masterplan without checking runtime boxes. | R27,R28; all units; Final links/IDs/status/diff audit; active pointer stays awaiting implementation instruction. |

These entries record the observed evidence, affected requirement/unit, correction
and required retest. Their common reproduction is reading the original Feature
005 outline against the live source and product contract; entries 003, 004, 006,
010 and 011 additionally use version-specific upstream source. No injected
runtime failure is represented as already observed.

## Open execution gates

| ID | Owner | Severity | Status / blocking effect | Next action | Affected scope / closure evidence |
| --- | --- | --- | --- | --- | --- |
| 005-ISS-021 | U1 owner | P1 | Open — blocks U7 | Prove combined graph/minimum Go, CLI formatter compatibility and license inventory through canonical targets. | R4; V01–V03,V12; Exact versions/compiler/source and passing local receipts; dependency selection is not build proof. |
| 005-ISS-022 | Each unit owner, U6 acceptance owner | P1 | Open — blocks Phase 5 local acceptance | Execute red/green/full/race/coverage and real disk/lifecycle cases, resolve findings, synchronize and commit each unit. | R1–R28; V01–V110; Eight unit commits, actual focused/aggregate receipts and zero unresolved local blockers. |
| 005-ISS-023 | Terminal verifier / U7 through U6 | P1 | Open — blocks local visual acceptance | Use owned Kitty plus isolated DB, inspect stated interactions; if desktop unavailable keep evidence pending with actual blocker. | R9,R12,R15,R22,R26; V26,V39,V56,V67,V86,V100,V104; Dated visible observations/captures, sizes, terminal and binary hash; headless/PTY evidence does not close it. |
| 005-ISS-024 | Performance owner / U6 | P1 | Open — blocks Phase 5 local acceptance | Run retained-sample CLI matrix and TUI preparation budgets with host metadata and no competing workload. | R25; V106–V108; All per-case/run CLI and preparation limits pass, all samples retained; stress/startup metrics labeled observations. |
| 005-ISS-025 | Feature 006 release maintainer | P1 | Deferred — blocks release, not local Phase 5 acceptance | Plan/execute native matrix and hosted packaging against final artifacts; preserve the handoff in all packs. | R4,R22,R26,R28; V111–V112; Native/hosted evidence and separately authorized release action; local builds are insufficient. |

### 005-ISS-021: Compatibility execution

- **Phase found:** Planning. **Severity:** P1. **Owner:** U1 owner.
- **Status / blocking effect:** Open — blocks U7.
- **Affected scope:** R4; V01–V03,V12.
- **Evidence now:** No execution receipt; the planning pass cannot establish this result.
- **Next action:** Prove combined graph/minimum Go, CLI formatter compatibility and license inventory through canonical targets.
- **Closure evidence:** Exact versions/compiler/source and passing local receipts; dependency selection is not build proof.
- **Revisit:** At U1 before first unit commit. This is an owned execution gate, not a failed test or an accepted waiver.

### 005-ISS-022: Local implementation and integrity

- **Phase found:** Planning. **Severity:** P1. **Owner:** Each unit owner, U6 acceptance owner.
- **Status / blocking effect:** Open — blocks Phase 5 local acceptance.
- **Affected scope:** R1–R28; V01–V110.
- **Evidence now:** No execution receipt; the planning pass cannot establish this result.
- **Next action:** Execute red/green/full/race/coverage and real disk/lifecycle cases, resolve findings, synchronize and commit each unit.
- **Closure evidence:** Eight unit commits, actual focused/aggregate receipts and zero unresolved local blockers.
- **Revisit:** At each unit boundary; final U6. This is an owned execution gate, not a failed test or an accepted waiver.

### 005-ISS-023: Owned Kitty acceptance

- **Phase found:** Planning. **Severity:** P1. **Owner:** Terminal verifier / U7 through U6.
- **Status / blocking effect:** Open — blocks local visual acceptance.
- **Affected scope:** R9,R12,R15,R22,R26; V26,V39,V56,V67,V86,V100,V104.
- **Evidence now:** No execution receipt; the planning pass cannot establish this result.
- **Next action:** Use owned Kitty plus isolated DB, inspect stated interactions; if desktop unavailable keep evidence pending with actual blocker.
- **Closure evidence:** Dated visible observations/captures, sizes, terminal and binary hash; headless/PTY evidence does not close it.
- **Revisit:** Each UI-bearing unit and final candidate. This is an owned execution gate, not a failed test or an accepted waiver.

### 005-ISS-024: CLI distribution and TUI measurement

- **Phase found:** Planning. **Severity:** P1. **Owner:** Performance owner / U6.
- **Status / blocking effect:** Open — blocks Phase 5 local acceptance.
- **Affected scope:** R25; V106–V108.
- **Evidence now:** No execution receipt; the planning pass cannot establish this result.
- **Next action:** Run retained-sample CLI matrix and TUI preparation budgets with host metadata and no competing workload.
- **Closure evidence:** All per-case/run CLI and preparation limits pass, all samples retained; stress/startup metrics labeled observations.
- **Revisit:** U6 after final binary changes. This is an owned execution gate, not a failed test or an accepted waiver.

### 005-ISS-025: Native and hosted proof

- **Phase found:** Planning. **Severity:** P1. **Owner:** Feature 006 release maintainer.
- **Status / blocking effect:** Deferred — blocks release, not local Phase 5 acceptance.
- **Affected scope:** R4,R22,R26,R28; V111–V112.
- **Evidence now:** No execution receipt; the planning pass cannot establish this result.
- **Next action:** Plan/execute native matrix and hosted packaging against final artifacts; preserve the handoff in all packs.
- **Closure evidence:** Native/hosted evidence and separately authorized release action; local builds are insufficient.
- **Revisit:** Feature 006 planning and exact-candidate release gate. This is an owned execution gate, not a failed test or an accepted waiver.

## Review-lens sign-offs

This is a **sequential planning review by the current agent**, following the
repository's sequential dispatch mapping. It is not independent peer or
cross-model corroboration. All selected lenses cover the three artifacts.

| Lens | Planning disposition | Findings/evidence | Implementation retest |
| --- | --- | --- | --- |
| Coherence and traceability | Reviewed; corrected outline/status/crosslinks | ISS-001,002,020; 28 R IDs / eight U IDs / 112 V IDs | Unit/status/source audit at each boundary |
| Architecture and feasibility | Reviewed; preserve ports and v1, split lifecycle/consent | ISS-003–009,017; code and pinned framework source | U1 graph and U7 drain/recovery proof |
| Product scope and agent parity | Reviewed; existing CLI covers durable actions | ISS-002,012; product R18–R26 and CLI grammar | U5/U8 workflows plus unchanged CLI JSON |
| Correctness and reliability | Reviewed; outcome before cause, no duplicate write or stale apply | ISS-007,009,013–017 | Barrier-driven, real disk and cancellation cases |
| Security/privacy boundary | Reviewed; task text is untrusted terminal input | ISS-010,011,016; source/output sanitization | Control/Markdown/panic payload fixtures |
| Design and accessibility | Reviewed; focus, exact geometry, raw text and consent | ISS-011,014,018; key/state/layout contracts | Synthetic geometry plus owned Kitty |
| Data integrity and concurrency | Reviewed; Base/Expected and owner retirement | ISS-006–009,012–017 | Atomic graph/history, two-owner conflicts and recovery |
| Simplicity and maintainability | Reviewed; no new schema/cache/framework | KTD2–KTD7; finite flags/tokens and constructor state | Scope/import review; avoid generic dispatcher expansion |
| Performance and portability | Reviewed; dependency/measurement proof still open | ISS-003,019,021,024,025 | Retained CLI/TUI samples, minimum-Go and native tiers |
| Test strategy and evidence quality | Reviewed; cold purity and separate proof tiers | ISS-004,019,022,023; matrix/fixtures | Red receipts, ≥95% coverage, full/race, terminal inspection |

Runtime compatibility, terminal quality and performance have evidence gates,
not invented assurances. The plan selects a mechanism for each; an implementation
finding that invalidates it must update all affected artifacts before proceeding.

## Unit workorder and commit boundaries

| Sequence | Unit / owner | Required completion before advancing |
| --- | --- | --- |
| 1 | U1 / root and compatibility owner | V01–V12, dependency receipt, make validate, synchronized pack/masterplan, local commit |
| 2 | U7 / runtime owner | V13–V26 including visible shell/lifecycle, canonical gate, synchronized local commit |
| 3 | U2 / presentation owner | V27–V39 and unchanged CLI sanitation, canonical gate, synchronized local commit |
| 4 | U3 / navigation owner | V40–V56 including external changes and timer bounds, canonical gate, synchronized local commit |
| 5 | U4 / details owner | V57–V67 including security/fallback/history, canonical gate, synchronized local commit |
| 6 | U5 / forms owner | V68–V86 including raw preservation and real disk conflicts, canonical gate, synchronized local commit |
| 7 | U8 / consent/recovery owner | V87–V100 including unknown-create quarantine and fresh-owner proof, canonical gate, synchronized local commit |
| 8 | U6 / acceptance owner | V101–V110, final-candidate earlier scenarios, latency/Kitty/review/docs, synchronized local commit |

If a mandatory unit terminal scenario is blocked, continue useful automated work
within that unit and record the exact blocker; do not check the unit complete
or advance its masterplan pointer. Do not stage unrelated changes. Each receipt
must identify the code it proves; a later production change invalidates affected
earlier acceptance until rechecked.

## Implementation release gate

Every checkbox below remains unchecked during planning.

- [ ] U1, U7, U2, U3, U4, U5, U8 and U6 implemented and separately committed.
- [ ] Red failures and focused green results retained for each feature-bearing change.
- [ ] Canonical make validate passes with minimum coverage and no exceptions.
- [ ] Minimum Go and five CGO-free application/test cross-builds pass.
- [ ] V01–V110 local scenarios executed with current source/candidate evidence.
- [ ] Required owned Kitty inspections complete; no headless substitution.
- [ ] CLI latency distributions and TUI preparation budgets pass with all samples.
- [ ] Local P0/P1/P2 findings resolved and rechecked; no unknown-outcome replay.
- [ ] docs/tui.md, README and product/feature/masterplan handoffs synchronized.
- [ ] V111–V112 native/hosted release proof completed by Feature 006.
- [ ] Publication explicitly authorized and performed only through its release gate.

Local Phase 5 completion may leave the last two items open with this explicit
Feature 006 ownership. It may not claim released, native-green, hosted-green or
all issues closed. No implementation, commit, push, PR or publication is part of
this plan-ultrathink invocation.

## Planning audit record

The final sequential coherence/feasibility pass tightened ISS-006/007/016:
recovery outranks a waiting write; a stale read's unknown outcome is handled
before its display payload is dropped; recovery must not drain its own active
command; and a recording output wrapper detects errors the renderer may ignore.
V24/V51/V94 explicitly cover those corrections. Search cancellation also
reconciles the restored selection against the current forest. These are plan
corrections, not observed runtime test results.
The CLI runner seam also separates acknowledged earlier changes from a later
unknown outcome. V21 prevents reusing a single-command committed flag as the
outcome of an entire multi-mutation TUI session (ISS-005/016).

2026-09-29: live baseline e899491; first-party artifacts remain under docs;
required links, R/U/V coverage, preserved original U IDs and planning-only
checkboxes audited. Source/contract review and document inspection only.
Application tests, make validate, actual dependency builds, benchmarks and
visible terminal scenarios were not run. No independent review is claimed.

### U1 execution checkpoint — 2026-09-29

U1 passes canonical validation and Go 1.25 short tests/five-target CGO-free
cross-builds; see [receipt](../verification-evidence/005/u1.json). ISS-021 has
initial compatibility proof; final linked-candidate proof remains U6. ISS-022
and other runtime gates remain open. V06/V10 evidence is scoped to current
widgets/wait primitive and must be extended by their later consumer units.
Next: U7, including actual owned Kitty session/lifecycle acceptance.

### U7 compatibility finding — eager terminal discovery (2026-09-29)

The first linked application exposes Bubble Tea v1.3.10 `tea_init.go`: its
package initializer calls global `lipgloss.HasDarkBackground` before CLI
admission. An unanswered OSC query delays startup by five seconds and breaks
existing real-PTY CLI confirmation tests. The new no-discovery regression is
Red; see `docs/verification-evidence/005/u7-red-global-terminal-query.log.gz`.

KTD1 is amended before remediation: retain the v1.3.10 APIs and module graph,
but use a repository-local copy under `third_party/bubbletea` with only that
initializer removed. Preserve the upstream MIT license, complete source/test
files, version/checksum provenance and per-file hashes. The TUI already owns
explicit per-session color/background settings. Do not set global renderer
state, require environment workarounds or relax timing/readiness gates.
Canonical module checks, source-integrity tests, the real PTY suite and Kitty
relaunch must pass before U7 can complete. This is a narrow compatibility patch,
not a v2 migration; U6 must repeat final-candidate CLI distribution checks.

### U7 local acceptance — 2026-09-29

[Lifecycle receipt](../verification-evidence/005/u7.json) records V13–V26,
red/green regressions, canonical validation, minimum-Go/race/cross-build proof
and real Kitty app usage. The pinned-source initializer patch is verified;
final CLI distributions remain U6. No test/build/benchmark results run in Kitty.
U2 bounded geometry and shared safe text is next.

### U2 local acceptance and runtime findings — 2026-09-29

[Receipt](../verification-evidence/005/u2.json) closes the U2 implementation
boundary with canonical/minimum-Go gates, unchanged CLI fixtures and real Kitty
inspection. The user-selected spacious rows and shadow correction are adopted.
Uniform cell backgrounds, grapheme clipping, plain focus and resize are proven.

Kitty found that the U7 recording writer hid `term.File`, so Bubble Tea never
received initial size or subsequent resize events. The failing child PTY resize
case and screenshot are retained; preserving the interface fixes the cause
without bypassing output-error recording. TERM/COLORTERM select available TUI
colors and NO_COLOR remains authoritative; no appearance query was introduced.

ISS-018's geometry/help/overlay foundation is accepted; form/consent and saving
portions remain with U5/U8. ISS-010's shared safe-text portion is accepted; U4
still owns Markdown. ISS-022/023 remain open for later units, with U2 evidence
complete. U3 navigation/search/refresh is next.
