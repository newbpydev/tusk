---
feature-id: "005"
plan-source: ../plans/2026-09-06-005-feat-interactive-tui-application-plan.md
verification-plan: ../verification-plans/2026-09-06-005-feat-interactive-tui-application-verification-plan.md
status: Locally accepted - all eight units complete; native and hosted release gates deferred
evidence-scope: Local V01–V110 and unit receipts; V111–V112 remain Feature 006
---

# Feature 005 Issue Workorder

### PR #5 reviewer startup repair

005-ISS-029 (P2): Kilo could not start reviewing PR #5 because Glamour's copied
gallery LFS pointers referenced objects absent from Tusk's LFS server. Workspace
checkout failed with exit 128. Materialize the eight exact upstream v0.9.1 PNGs
as regular Git blobs and disable their LFS filters; preserve original pointer
hashes and declare the nine checkout changes in the source manifest. Renderer
patches and app behavior remain unchanged.

The [startup repair evidence](../verification-evidence/005/pr5-reviewer-startup/README.md)
records the checkout regression Red/Green and all eight verified image hashes.
Canonical validation is required before commit. Closure requires publication,
a fresh remote checkout with LFS unavailable and Kilo passing workspace setup;
the later review result and Feature 006 release gates remain separate.
ISS-029 is closed by published repair commit `4236d83`, its passing fresh remote
clone check and Kilo's observed review skill/PR-reading activity on that head.

The [plan](../plans/2026-09-06-005-feat-interactive-tui-application-plan.md) and
[verification plan](../verification-plans/2026-09-06-005-feat-interactive-tui-application-verification-plan.md) are one
planning pack. Work only in the sequence authorized by [MASTERPLAN.md](../../MASTERPLAN.md).

**Current result:** the 20 planning findings have local implementation evidence
through V01–V110 and the eight unit receipts. ISS-021–ISS-024 are closed locally;
ISS-025 remains a Feature 006 release obligation. The historical planning register
below preserves its original findings and planned retests. Current acceptance is
recorded in [U6 acceptance](../verification-evidence/005/u6-acceptance.json).

## Historical planning issue register

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
| 005-ISS-011 | Planning/source review | Data preservation / U5 owner | P1 | Closed locally in U5; u5.json | Widget sanitization/line limits can silently rewrite original notes during metadata edits. Keep raw values and dirty fields; require lossless round-trip or explicit read-only replacement path; reject paste atomically. | R23; U5; Bubbles source; V33,V58,V70,V84–V85. |
| 005-ISS-012 | Planning/source review | Parity / U5 owner | P1 | Closed locally in U5; u5.json | Forms could omit clear intent, misuse due display values or recompute business rules. Map only dirty fields to existing Base/clear/patch APIs; service owns graph/status/progress. | R16–R18; U5; ports.TaskService and service.md; V68–V80. |
| 005-ISS-013 | Planning/source review | Reliability / U5 owner | P1 | Closed locally in U5; u5.json | Duplicate submit or saved-refresh-failed state can repeat an acknowledged mutation. One admission before waiting for read drain; known success closes draft, stale read blocks writes. | R19; U5; V81–V83,V20. |
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

## Execution and release gates

| ID | Owner | Severity | Status / blocking effect | Next action | Affected scope / closure evidence |
| --- | --- | --- | --- | --- | --- |
| 005-ISS-021 | U1 owner | P1 | Closed locally — U1 receipt; final U6 minimum-Go proof retained | Prove combined graph/minimum Go, CLI formatter compatibility and license inventory through canonical targets. | R4; V01–V03,V12; Exact versions/compiler/source and passing local receipts; dependency selection is not build proof. |
| 005-ISS-022 | Each unit owner, U6 acceptance owner | P1 | Closed locally — U6 acceptance receipt and containing commit | Retain eight unit boundaries, red/green/full/race/coverage, disk/lifecycle and review evidence. | R1–R28; V01–V110; u6-acceptance.json; zero unresolved local blockers. |
| 005-ISS-023 | Terminal verifier / U7 through U6 | P1 | Closed locally — owner approved 2026-09-29 | Use owned Kitty plus isolated DB, inspect stated interactions; if desktop unavailable keep evidence pending with actual blocker. | R9,R12,R15,R22,R26; V26,V39,V56,V67,V86,V100,V104; Dated visible observations/captures, sizes, terminal and binary hash; headless/PTY evidence does not close it. |
| 005-ISS-024 | Performance owner / U6 | P1 | Closed locally — 84/84 CLI and 78/78 TUI case-runs pass | Retain all samples, earlier failed reports and declared host conditions. | R25; V106–V108; cli-latency.json, tui-latency.json, u6-acceptance.json; stress/startup metrics remain observations. |
| 005-ISS-025 | Feature 006 release maintainer | P1 | Deferred — blocks release, not local Phase 5 acceptance | Plan/execute native matrix and hosted packaging against final artifacts; preserve the handoff in all packs. | R4,R22,R26,R28; V111–V112; Native/hosted evidence and separately authorized release action; local builds are insufficient. |

### 005-ISS-021: Compatibility execution

- **Phase found:** Planning. **Severity:** P1. **Owner:** U1 owner.
- **Status / blocking effect:** Closed locally in U1, reconfirmed by final U6 minimum-Go checks.
- **Affected scope:** R4; V01–V03,V12.
- **Evidence now:** u1.json and u6-acceptance.json retain combined graph, Go 1.25 and five-target build proof.
- **Next action:** Prove combined graph/minimum Go, CLI formatter compatibility and license inventory through canonical targets.
- **Closure evidence:** Exact versions/compiler/source and passing local receipts; dependency selection is not build proof.
- **Revisit:** At U1 before first unit commit. This is an owned execution gate, not a failed test or an accepted waiver.

### 005-ISS-022: Local implementation and integrity

- **Phase found:** Planning. **Severity:** P1. **Owner:** Each unit owner, U6 acceptance owner.
- **Status / blocking effect:** Closed locally by U6 acceptance and its containing commit.
- **Affected scope:** R1–R28; V01–V110.
- **Evidence now:** All eight unit receipts, final canonical validation, 110 local scenarios and completed review receipts are bound in u6-acceptance.json.
- **Next action:** Execute red/green/full/race/coverage and real disk/lifecycle cases, resolve findings, synchronize and commit each unit.
- **Closure evidence:** Eight unit commits, actual focused/aggregate receipts and zero unresolved local blockers.
- **Revisit:** At each unit boundary; final U6. This is an owned execution gate, not a failed test or an accepted waiver.

### 005-ISS-023: Owned Kitty acceptance

- **Phase found:** Planning. **Severity:** P1. **Owner:** Terminal verifier / U7 through U6.
- **Status / blocking effect:** Closed locally; Spacious design and calendar explicitly approved by the owner.
- **Affected scope:** R9,R12,R15,R22,R26; V26,V39,V56,V67,V86,V100,V104.
- **Evidence now:** u6-kitty/README.md and u6-calendar.md retain actual app/CLI interactions, screenshots and terminal restoration evidence.
- **Next action:** Use owned Kitty plus isolated DB, inspect stated interactions; if desktop unavailable keep evidence pending with actual blocker.
- **Closure evidence:** Dated visible observations/captures, sizes, terminal and binary hash; headless/PTY evidence does not close it.
- **Revisit:** Each UI-bearing unit and final candidate. This is an owned execution gate, not a failed test or an accepted waiver.

### 005-ISS-024: CLI distribution and TUI measurement

- **Phase found:** Planning. **Severity:** P1. **Owner:** Performance owner / U6.
- **Status / blocking effect:** Closed locally; unchanged limits pass on the declared Konsole-session host.
- **Affected scope:** R25; V106–V108.
- **Evidence now:** 84/84 CLI and 78/78 TUI case-runs pass; all samples, target misses and earlier failed reports remain retained.
- **Next action:** Preserve receipts; remeasure affected behavior if later production code changes.
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

Current local implementation evidence is complete; native/hosted release gates remain separate.

- [x] U1, U7, U2, U3, U4, U5, U8 and U6 implemented and separately committed.
- [x] Red failures and focused green results retained for each feature-bearing change.
- [x] Canonical make validate passes with minimum coverage and no exceptions.
- [x] Minimum Go and five CGO-free application/test cross-builds pass.
- [x] V01–V110 local scenarios executed with current source/candidate evidence.
- [x] Required owned Kitty inspections complete; no headless substitution.
- [x] CLI latency distributions and TUI preparation budgets pass with all samples.
- [x] Local P0/P1/P2 findings resolved and rechecked; no unknown-outcome replay.
- [x] docs/tui.md, README and product/feature/masterplan handoffs synchronized; registry and final local acceptance reconciled on 2026-09-30.
- [ ] V111–V112 native/hosted release proof completed by Feature 006.
- [ ] Publication explicitly authorized and performed only through its release gate.

Local Phase 5 completion may leave the last two items open with this explicit
Feature 006 ownership. It may not claim released, native-green, hosted-green or
all issues closed. The owner-authorized ce-work execution closes the local gates;
push, PR and publication remain separately authorized actions.

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

### U3 local acceptance — 2026-09-29

[Navigation receipt](../verification-evidence/005/u3.json) records complete-forest
groups, spacious tree rows, selection by ID/incarnation, temporary search expansion,
150 ms cancellable debounce, status/priority/tag/due filters and coalesced refresh.
The due-day parser is injected from the composition root and resolves relative
expressions only on Apply. Root View still returns its prepared immutable frame.

Canonical validation, Go 1.25 race checks and five CGO-free target builds pass.
A real two-owner SQLite test and owned Kitty CLI/TUI interactions cover external
rename, priority change, reparent and deletion while searching. Live inspection
also verifies collapse restoration, filter draft shrink/restore, 80x24 controls,
NO_COLOR focus and clean terminal restoration. No tests/builds ran in Kitty.

V40–V49 and V56 close here. V50–V55 retain their later consumer portions:
post-write refresh and write admission (U5), history dispatch (U4), edit Base and
form identity (U5), and fresh-owner recovery UI (U8). The current scheduler
releases matching busy slots and freezes unknown outcomes before rejecting stale
payloads; U8 must connect its Reload/readback UI before feature acceptance.
U4 details, Markdown and timeline is next. No mutation UI is claimed yet.

### U4 renderer budget finding — 2026-09-29

A 1.0625 MiB unbroken Unicode note spent 146 seconds in Glamour's reflow
word-wrapper before the owned test process was stopped with a retained stack
trace. KTD9 is refined before integration: Markdown formatting is admitted only
for at most 32 KiB of sanitized source with no whitespace-delimited token above
2 KiB. Larger/pathological notes use labeled, safely wrapped plain text in the
same asynchronous render lane. Every byte of raw notes remains untouched and
all displayed text remains scrollable. This is a formatting budget, not a storage
or editing limit. V58/V62 must prove full content and bounded pending work; U6
retains final preparation/render measurements. U4 remains in progress.

### U4 local acceptance — 2026-09-29

[Details receipt](../verification-evidence/005/u4.json) records complete wrapped
metadata, one scrollable notes/activity viewport, fixed workspace Markdown
styling, generated-output filtering and real ascending int64 service events.
History uses the single service-operation slot; Markdown has one active render
and one latest pending request. Task incarnation, content, width and profile
reject obsolete render replies; history freshness remains independent.

Canonical validation (TUI coverage 98.6%), Go 1.25 affected race checks and all
five CGO-free builds pass. Kitty verifies 80x24/120x40/200x60, monochrome Markdown,
selection changes, real history and long-note Home/End/page scrolling. Live
inspection added contextual scrolling hints and preserved End intent through
pending rendering and resize. The 1.0625 MiB test uses complete plain fallback
under the documented formatting budget; no raw note is truncated or rewritten.

V35,V54,V57–V64,V66–V67 close. V65 known/missing/unknown read handling passes;
its user-facing fresh-owner Reload/readback completion remains U8. Existing
V50–V53,V55 and form/consent presentation portions stay with U5/U8. U5 forms and
nondestructive mutations is next. No production mutation UI is claimed yet.

### U5 editor capacity contract — 2026-09-29

The form editor admits at most 64 KiB per text field and at most 10,000 notes
lines. This bounds widget preparation only; larger existing stored values remain
raw and read-only, with explicit default-Cancel replacement and omission from
unrelated patches. Before loading a widget, exact round-trip equality is required.
Incoming paste is validated atomically before mutation; notes allow LF, while
controls, bidi overrides and decoded replacement-rune input are rejected. The
replacement rune is conservatively unsupported on input because terminal key
messages cannot distinguish it from already-decoded invalid UTF-8. Existing
valid stored values still follow the round-trip preservation rule. V84/V85 prove
these boundaries; no storage/schema limit or automatic truncation is introduced.


### U5 interaction refinements — 2026-09-29

Below 80×24, hidden forms ignore ordinary input and retain their exact draft and
focus until restored; Ctrl+C remains available and browse q still quits. This
refines the earlier below-minimum wording that allowed invisible draft typing.
Form headers show field position, and the due label states the configured zone.
Saved-refresh failure explicitly says writes are paused; an admitted mutation
says Saving rather than Refreshing. Help and quit remain visible with notices at
80 columns. The parent picker accepts literal spaces in multiword searches.

The manual `make build-tui-fixture` target compiles an application-only launcher
from test source in Bash. Its Kitty entry point runs the production Run/model,
real SQLite service and a controlled post-create read failure, then exits before
the test runner prints results. It is fault-injected application evidence, not
a test-results display or proof that an unmodified storage fault occurred.


### U5 local acceptance — 2026-09-29

U5 passes `make validate build check-generated`, minimum-Go affected race checks
and five CGO-free application/test builds. TUI coverage is 98.4%. The
[U5 receipt](../verification-evidence/005/u5.json) retains source/binary hashes,
red/green observations, logs, real disk contracts and owned Kitty captures.
Create/edit forms preserve raw fields and detached Base, reject paste atomically,
and send only changed fields. Read cancellation drains before one admitted write;
known failures retain drafts, conflicts require explicit reload, and committed
writes cannot be replayed after failed readback. Moves/rollups and lifecycle
policy use the service unchanged. Kitty confirms 80×24/120×40/200×60, undersize
restoration, NO_COLOR, multiline paste, parent selection, conflict/discard, CLI
JSON persistence and the controlled saved-refresh-failed screen.

V38,V50,V52,V53,V68–V86 close. V30/V36/V51/V55 retain their U8 consent/recovery
portions. Product TUSK-V62 closes; TUSK-V61/V63 retain deletion/consent portions.
Inline contract, integrity, security, async, usability and simplification review
resolved the recorded findings; no independent reviewer claim is made. U8 is
next. Performance/final acceptance stays U6; native/hosted release stays Feature 006.

Final Kitty readback review caught an obsolete failure notice after successful
refresh. The red/green correction and fresh capture are retained in u5.json;
canonical and minimum-Go gates passed again before the unit commit.


### U8 local acceptance — 2026-09-29

[U8 evidence](../verification-evidence/005/u8.json) records passing canonical
validation (96.8% TUI coverage), minimum-Go race/cross-build checks, real disk
consent/recovery tests and inspected Kitty color/plain screens. Deletion defaults
to Cancel, pins the exact preview and resets recursion after membership conflict.
Unknown outcomes freeze writes, retain read-only intent, close the old owner
before readback, and require acknowledgment before independent new actions.
No mutation is replayed or uncertain create identity guessed. A later known
commit does not erase the earlier unknown receipt. Close failure permits only
quit; old-owner messages cannot publish. All overlay background tests pass.

V30,V36,V51,V55,V65 and V87–V100 close. Cancellation combines the U7 raw-key
receipt tests and U8 exclusive-owner drain barriers; final child-PTY proof stays
U6. Controlled service-outcome injection in Kitty is labeled separately from
real SQLite transaction evidence. Inline reviews resolved the recorded findings;
no independent review is claimed. U6 now owns final workflow, performance,
documentation and review, including long browse notices/action-specific toggle
errors and measured large-preview/projection costs. Native/hosted release proof
remains Feature 006.


### U6 presentation refinement checkpoint (2026-09-29)

The owner's UI/UX steering makes visual fidelity part of local acceptance.
The Spacious reference now governs persistent search and working All/Today/Done
tabs, status symbols, consistent modal insets, aligned input rules, checkboxes,
buttons and focus markers. Task details put everyday information and notes before
audit metadata. Today follows local midnight; custom date filters remain fixed.
Raw drafts, exact deletion consent, single-operation admission and recovery
barriers are unchanged. Canonical validation, minimum-Go tests/five-target builds
and focused visual regressions pass. Owned Kitty inspection covers colored/plain
controls, 80x24/120x40 and large layouts; final performance, review and remaining
terminal acceptance are still in progress. U6 remains the active target.

### U6 CLI startup regression and correction (2026-09-29)

The first final CLI matrix failed all 84 case-runs. The retained report is
`docs/verification-evidence/005/u6-logs/cli-latency-eager-markdown.json`;
`GODEBUG=inittrace=1` identified Chroma's eager language and theme XML loading
as the dominant cost before command routing (about 11 ms on this host).
Glamour remains pinned at v0.9.1, with two local renderer
patches: base-styled code blocks without Chroma, and HTML text extraction using
the existing HTML tokenizer without CSS/URL policy registries. Markdown headings,
lists, tables, safe links and code text remain supported; tags/attributes and
script/style contents are omitted and decoded text crosses the terminal sanitizer;
code blocks use uniform coloring rather than language-specific highlighting.
The source/license copy and SHA-256 manifest live in `third_party/glamour`;
`TestCLIStartup_NoSyntaxRegistryInitialization` is Red before the patch and
Green after it. The full unmodified timing policy must still pass before U6
acceptance. This is a measured startup correction, not a waived latency gate.

The syntax-only correction passed 56/84 CLI case-runs but retained help/version
and large-query misses (`u6-logs/cli-latency-no-syntax.json`). The second patch
removes the remaining sanitizer registry initialization; HTML text/markup/entity
characterization passes, and the startup regression remains Red-to-Green.

### U6 owner approval and remaining latency gate (2026-09-29)

The owner approved the refined Spacious UI: “Yes, keep this direction.” Padding,
quick tabs, checkboxes and opaque dialog backgrounds are the accepted visual
baseline. Final owned Kitty checks cover 80×24, 120×40 and exact 200×60, help
paging, terminal restoration, CLI usage, relaunch and injected recovery. See
[u6-kitty receipt](../verification-evidence/005/u6-kitty/README.md).

The two startup corrections pass canonical validation and Go 1.25 tests/builds.
The retained CLI report improves to 74/84 passing case-runs but still fails the
unchanged gate. A same-host Feature 004 control is being measured to distinguish
startup overhead from desktop contention. No failed samples are removed. U6,
V106 and Phase 5 acceptance remain open; final review and measurements must be
settled before the unit commit.

### U6 review follow-up (2026-09-29)

The full-branch ce-code-review round is complete. Local lenses ran sequentially
inline per AGENTS. Claude returned a provider authentication failure; the Grok
replacement through Cursor returned four observations, with serving identity
unverified and no independent-agreement promotion. Three claims were rejected
against explicit existing contracts; the useful production-process test gap was
closed by TestTUIProcess_CommittedBeforeInterrupt. It creates a real task using
the production binary, then verifies SIGINT restoration, committed receipt and
fresh CLI readback. Focused, canonical and minimum-Go checks pass. See
[review receipt](../verification-evidence/005/review-r1/review.json) and its
[follow-up](../verification-evidence/005/review-r1/follow-up.json).

A same-host Feature 004 control passed 82/84 cases; a later current run passed
77/84 while system package updates were observed consuming CPU. Reports and
host observations are retained. The update has ended; one separate current
matrix is running with no owned test/build/review workload. The limits and
fixtures remain unchanged. This is diagnosis of changed host conditions, not
selection of individual passing runs or deletion of outliers.

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

### U6 owner-requested tree and progress refinement (2026-09-30)

The owner authorized this follow-up after local acceptance at 03c5381. R9/R10/R12
and R13/R18 retain their contracts while presentation improves: draw connected
visible sibling branches with ancestor continuations through spacious metadata
rows; only the last visible sibling receives an elbow. Search, sorting, collapse,
scrolling and narrow layouts must preserve truthful visible connections.

Progress displays equivalent completed direct items: leaf progress / 100 out of
one; a parent sums its immediate children's authoritative progress / 100 out of
its child count. Thus a half-complete child plus two untouched siblings displays
(0.5 / 3). Keep the stored percentage and bar; do not recompute domain rollups,
change status, or treat filtered/collapsed children as absent. Integers have no
decimal suffix; fractional values retain up to two meaningful decimal places.

This is a U6 follow-up, not a new phase or unit identity. Red-first checks cover
nested/sorted/filtered/collapsed branches, continuous metadata guides, all/none/
partial and single-task progress, complete snapshots despite filters, resize,
View purity and real-disk refresh after descendant completion. Run canonical
validation, affected minimum-Go checks, retained CLI/TUI measurements, code
review and actual owned Kitty checks at 80x24 and 120x40 in color/plain modes.
Owner visual confirmation and a coherent local follow-up commit close this
refinement. Earlier acceptance receipts remain unchanged; release proof stays
with Feature 006. Current refinement status: locally accepted; final receipt below.

### Darkmatter palette refinement (2026-09-30)

The owner approved the connected branches and fractional progress display, then
requested the Darkmatter palette at https://tweakcn.com/editor/theme?theme=darkmatter.
Use the dark preset from jnsahaj/tweakcn commit
a3b47b37cba97dd637de517aab52c45ec0f83456 (utils/theme-presets.ts): neutral near-black,
soft gray text, warm orange focus/actions, gray selection and peach warnings.
Adapt the preset's low-contrast border to its accent gray for terminal strokes;
use readable muted text for inactive panel titles and full foreground text for
footer shortcuts. Preserve red destructive warnings using the preset's light
destructive color, because its dark destructive token is teal. Keep all surfaces
opaque, existing non-color focus markers and the approved spacing/geometry.
The owner confirmed “Yes, keep Darkmatter”; the tree/count approval is also
recorded. Canonical, affected Go 1.25, owned Kitty and all 78 TUI timing cases
pass. The initial 80/84 CLI checkpoint is preserved below as historical
evidence; the final acceptance section records the resolved timing gate.

### Historical U6 visual follow-up checkpoint (2026-09-30, 17:59 UTC)

The owner approves the connected task guides, fractional progress counts and
Darkmatter palette. Red/green, canonical validation, affected Go 1.25 checks,
owned Kitty color/plain inspection at 80×24 and 120×40, scoped code review and
all 78 TUI timing cases pass. Stored progress semantics remain unchanged.

Final CLI acceptance remains open: the candidate passed 80/84, the accepted
03c5381 source control passed 81/84, and a subsequent candidate passed 77/84.
All samples and host observations are retained in the
[checkpoint](../verification-evidence/005/u6-hierarchy/checkpoint.json).
After Kilo activity fell, Chrome stayed near two cores across four observations;
no further identical-condition run is being repeated. The owner is being asked
for a quieter window or to retain the timing gate explicitly pending.

This is approved implementation with incomplete final acceptance. U6 remains
the active follow-up; Feature 006 implementation and publication are not started.

### U6 hierarchy and Darkmatter follow-up accepted (2026-09-30)

The owner approved the final app checks and reported Cline CLI shut down.
The unchanged application source now passes all 84 CLI timing cases across
three complete runs: worst query/help p90 14.291/4.112 ms; all tail guards
pass. Every sample, warmup and earlier failed report remains retained. Limits,
host settings and application behavior were not changed to obtain acceptance.

Fresh `make validate build check-generated` passes. The reviewed source hashes
still match the affected Go 1.25 checks, owned Kitty inspections and all 78
passing TUI timing cases (worst preparation p95 13.929 ms). The latter are
unchanged-source evidence, not a newly executed TUI matrix. The owner-approved
Spacious layout, calendar, Darkmatter palette, connected guides and fractional
direct-item counts are complete; product R7 stored rollup semantics are unchanged.

Implementation commit 246db8c and this containing acceptance commit close the
U6 follow-up and restore Phase 5 local acceptance. See the
[final receipt](../verification-evidence/005/u6-hierarchy/acceptance.json).
Feature 006 planning awaits instruction. V111/V112, remaining native TUSK-V66
coverage and hosted release proof remain separate; nothing was published.

### Owner-requested full-branch code review (2026-09-30)

The owner requested systematic P0-P2 remediation and repeated review to a clean
result. Three sequential main-context review rounds confirmed and corrected two
P2 defects: relative-date Save/Apply used a stale display clock at midnight, and
the patched-dependency integrity gate did not reject additional source files.
A P3 glossary correction aligns model ownership and context-specific keys with
the implemented UI. Stable review findings #1-#3 and discriminating red/green
receipts are retained in the [review receipt](../verification-evidence/005/review-local/acceptance.json).

Supplemental checks cover R4/R11/R16/R17/R27/R28 and V01-V03/V47/V69-V71/V105-V110:
`today`/`tomorrow` edits preserve the detached Base and dispatch the natural
expression after midnight; due filters select the current 23-hour DST day;
added source in either patched module fails the actual integrity guard. Final
`make validate build check-generated`, Go 1.25 focused race tests and five-target
cross-builds pass. Actual owned Kitty checks confirm calendar, relative due-date
Save/filter, approved styling and terminal restoration using an isolated DB.

The complete final CLI matrix passes 84/84 case-runs: worst query/help-version
p90 14.554/3.953 ms, with every tail guard passing. Earlier candidate and unchanged
control matrices both passed 82/84; both failed reports and all samples remain
retained. The final TUI matrix passes 78/78 case-runs, with worst preparation p95
13.798 ms and all 326 recorded source hashes matching the final source. Prior
date-fix-only measurements remain historical. No limit, sample or
runtime implementation was changed to obtain the final CLI pass.

The clean repeat review is local evidence, not independent corroboration.
AGENTS.md requires sequential main-thread review; the requested cross-model
Claude route failed HTTP 401 authentication and produced no verified peer
review. An inline adversarial pass completed instead. Native/hosted V111-V112 and
005-ISS-025 remain Feature 006 obligations. The checkout was already dirty, so
ce-code-review leaves the verified fixes uncommitted and preserves earlier
AGENTS.md/solution edits. No new implementation unit, publication or release is
claimed by this follow-up.

#### Review remediation register

| ID | Severity | Status | Defect and closure |
| --- | --- | --- | --- |
| 005-ISS-026 | P2 | Closed locally; containing review-fix commit | Save/Apply sampled the cached display clock and could drop a relative due-date edit or select yesterday. Sample the action clock; today/tomorrow midnight and DST filter regressions pass. Review #2; R11/R16/R17; V47/V69-V71. |
| 005-ISS-027 | P2 | Closed locally; containing review-fix commit | Patched-module hash checks overlooked added source. Check the actual inventory and declared patch names; isolated child tests reproduce and reject added files in both modules. Review #1; R4; V01/V105. |
| 005-ISS-028 | P3 | Closed locally; containing review-fix commit | CONCEPTS described immutable Update and global form keys that contradict production. Correct UI ownership, prepared View and contextual routing. Review #3; R27/R28; V110. |

These are verified corrections within the completed Feature 005, not additional
implementation units. At review return, ce-code-review left them uncommitted
under its dirty-tree rule. The owner's later commit request closes that boundary;
the approved compound documentation and verified review fixes are separate commits.

### Owner-authorized local commit closure (2026-09-30)

The owner requested committing the pending changes after the clean review.
Approved compound guidance is committed separately as dc6cd9f. The containing
review-fix commit records ISS-026 through ISS-028, their regression tests and
evidence, and the synchronized planning pack. Fresh `make validate` passes
before each commit. The original review report and acceptance receipt retain
their review-return snapshot; the [commit closure](../verification-evidence/005/review-local/commit-closure.json)
records the later authorization, quality gates and source identity. No runtime
source changed after review, so accepted timing and Kitty evidence still match.
Feature 006 native/hosted release obligations remain separate; nothing was
pushed or published.

### PR #5 hosted review R1 (2026-10-01)

Three stale-write UX findings are fixed with observed failing regressions:
paused form saves show an in-modal Ctrl+R hint; a paused delete supports r and
Ctrl+R, revokes old consent and loads a fresh preview; rejected toggles pause
writes until successful readback. Canonical `make validate build` passes.
Owned Kitty inspection at 100x32 confirms visible stale-form feedback and
retained raw draft. Delete-read abandonment and renewed consent are automated
regressions, not additional Kitty evidence. Receipts: `docs/verification-evidence/005/review-pr5/`.
Remaining hosted threads and Feature 006 native/release gates stay open.
