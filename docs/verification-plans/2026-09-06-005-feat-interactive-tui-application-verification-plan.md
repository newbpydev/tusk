---
feature-id: "005"
plan-source: ../plans/2026-09-06-005-feat-interactive-tui-application-plan.md
surface-profiles: [cli-tui, internal-service-consumer, persistence-lifecycle, documentation]
status: Locally accepted - V01–V110 complete; V111–V112 deferred to Feature 006
evidence-scope: All eight units locally accepted; native and hosted release proof separate
---

# Feature 005 Verification Plan

## Verification contract

The [implementation plan](../plans/2026-09-06-005-feat-interactive-tui-application-plan.md) defines behavior and units;
the [workorder](../workorders/2026-09-06-005-feat-interactive-tui-application-issues-workorder.md) owns unresolved findings
and evidence gates. There are **28 requirements, eight units and 112 planned
scenarios**. V01–V110 are local Feature 005 obligations; V111–V112 are Feature 006
release handoffs. No scenario was executed during this planning pass.

Supported implementation baseline: Go 1.25.0, selected Bubble Tea v1 family,
embedded SQLite and existing CLI/service contracts. Test Linux locally and
cross-build linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64.
Native and hosted proof is separate. Tests must use isolated temporary databases,
never the user's task database or existing terminal sessions.

Normal, boundary, malformed, failure, recovery, concurrency, compatibility and
non-regression cases below are obligations, not suggestions. A test named in the
plan is planned unless its file already exists. Every unit starts with failing
behavior, retains the failure, implements the narrow fix, reruns its focused
cases, then passes canonical validation before its separate commit.

---

## Requirement coverage

| Requirement | Owning/proving units | Scenarios |
| --- | --- | --- |
| R1 | U1, U7, U6 | V04, V13, V14, V26, V101, V106 |
| R2 | U1, U7, U2 | V09, V14, V15, V16, V26, V35 |
| R3 | U7, U6 | V17, V18, V19, V20, V21, V22, V24, V25, V26, V102, V103 |
| R4 | U1, U6, Feature 006 | V01, V02, V03, V12, V105, V106, V111 |
| R5 | U1, U2, U6 | V04, V05, V06, V08, V09, V37, V107 |
| R6 | U1, U7, U6 | V04, V07, V08, V10, V17, V25, V102, V108 |
| R7 | U1, U7, U2, U3 | V11, V17, V29, V52 |
| R8 | U1, U3, U6 | V10, V46, V50, V51, V52, V53, V54, V56, V101, V102, V108 |
| R9 | U2, U6 | V27, V28, V29, V30, V31, V37, V38, V39, V104, V107 |
| R10 | U2, U3, U6 | V31, V40, V41, V42, V43, V55, V56, V107 |
| R11 | U3 | V44, V45, V46, V47, V48, V49, V55, V56 |
| R12 | U2, U3, U5, U6 | V30, V36, V38, V39, V42, V55, V68, V72, V73, V86, V104 |
| R13 | U4, U6 | V57, V58, V59, V60, V61, V62, V66, V67, V107, V108 |
| R14 | U4 | V63, V64, V65, V66, V67 |
| R15 | U1, U2, U4, U6 | V03, V09, V31, V32, V33, V34, V35, V36, V39, V59, V60, V67, V104 |
| R16 | U5, U6 | V68, V69, V70, V71, V74, V86, V101 |
| R17 | U3, U5, U6 | V53, V73, V74, V75, V76, V82, V86, V101, V102 |
| R18 | U5, U8, U6 | V77, V78, V79, V80, V86, V92, V101 |
| R19 | U7, U3, U4, U5, U8 | V20, V51, V66, V71, V81, V82, V83, V86, V96 |
| R20 | U8, U6 | V87, V88, V89, V90, V91, V92, V99, V100, V101, V102 |
| R21 | U7, U8, U6 | V21, V22, V93, V94, V95, V96, V97, V98, V99, V100, V101, V102 |
| R22 | U1, U7, U2, U8, U6, Feature 006 | V10, V18, V19, V20, V21, V23, V24, V25, V26, V38, V98, V101, V103, V104, V111 |
| R23 | U1, U2, U4, U5 | V06, V33, V58, V70, V84, V85, V86 |
| R24 | U1, U2, U3, U4, U5, U8, U6 | V07, V28, V43, V46, V51, V53, V54, V56, V61, V64, V76, V89, V99, V101 |
| R25 | U1, U6 | V03, V106, V107, V108 |
| R26 | U1, U7, U2, U3, U4, U5, U8, U6, Feature 006 | V02, V26, V39, V56, V67, V86, V100, V103, V104, V105, V110, V111, V112 |
| R27 | U1, U6 | V08, V12, V105, V110 |
| R28 | U6, Feature 006 | V109, V110, V112 |

Product mapping: U16/TUSK-V54–V55 → U1/U2; U17/TUSK-V56–V58 → U3;
U18/TUSK-V59–V60 → U4; U19/TUSK-V61–V63 → U5/U8;
U20/TUSK-V64–V66 → U6. U7 supplies shared runtime proof for all of them.
Product checkboxes remain unchanged until execution establishes their own scope.

---

## Scenarios

### Root and compatibility — U1

- [x] V01 (R4; Compatibility): Compile against the selected v1 Model/KeyMsg/WindowSizeMsg APIs; resolve Bubble Tea 1.3.10, Bubbles 0.21.0, Lipgloss 1.1.0, Glamour 0.9.1 and ansi 0.10.1 without changing SQLite/libc pins.

- [x] V02 (R4,R26; Compatibility): Use Go 1.25.0 for tests and five CGO-free application/TUI test cross-builds; record resolved modules. Do not label non-host binaries executed.

- [x] V03 (R4,R15,R25; Regression): Run existing CLI width/control/JSON fixtures against the combined module graph; bytes and exit contracts remain unchanged despite the ansi upgrade.

- [x] V04 (R1,R5,R6; Boundary): Construct Model and call Init with fail-on-use service/clock/terminal/file fakes; construction prepares state, Init only returns commands, and no storage or terminal query occurs.

- [x] V05 (R5; Purity): Snapshot cold model before its first View; call View 100 times without warmup; compare exact frame and deep reachable state, including component caches.

- [x] V06 (R5,R23; Purity): Repeat the purity check with populated multiline textarea, help, loading, errors, selection, history and a modal; no pointer/map/slice/style/cache writes or commands.

- [x] V07 (R6,R24; Concurrency): Create a command, then change model/draft/selection and execute it; captured immutable inputs remain the original detached values and replies do not mutate the live model.

- [x] V08 (R5,R6,R27; Architecture): Review imports and constructor-owned state: no core/service UI imports, storage/CLI imports in tui, mutable package state, View callbacks or model worker goroutines.

- [x] V09 (R2,R5,R15; Isolation): Construct two sessions with different injected profiles/zones; their frames/options stay independent and no global Lipgloss/background/environment state leaks.

- [x] V10 (R6,R8,R22; Cancellation): Cancel a pending injected timer/debounce wait; it produces no new service work, duplicate timer chain or retained running wait.

- [x] V11 (R7; States): Drive initial loading, empty, loaded and known load failure messages with fake data; distinguish messages without changing outer dimensions or fabricating success.

- [x] V12 (R4,R27; Gate): Retain dependency/license inventory and canonical validation with no new coverage exemption; absent pins/failed module tidy or <95% package coverage prevents U1 completion.

### Invocation and session ownership — U7

- [x] V13 (R1; Grammar): Exercise tui, tui --help, help tui, extra args, unknown flags and --json; only valid explicit tui reaches its runner; syntax exits 2.

- [x] V14 (R1,R2; Regression): Bare invocation/help/version and tui help under invalid env, non-TTY and unwritable DB path return their ordinary help output with zero open/mkdir calls.

- [x] V15 (R2; Admission): Try stdin-only TTY, stdout-only TTY, neither, TERM=dumb and valid TTY streams; invalid terminals exit 1 before open/raw mode; stderr need not be a TTY.

- [x] V16 (R2; Configuration): Verify timezone and parent-policy flag > env > default and DB path precedence, including explicit false, invalid values and relative paths; recovery uses resolved initial identity.

- [x] V17 (R3,R6,R7; Lifecycle): Successful command-based open loads once; factory partial-open failure retains factory cleanup responsibility; nil service/closer is refused safely.

- [x] V18 (R3,R22; Lifecycle): Normal quit and failed Program startup restore the terminal and close each successfully opened owner exactly once; open failure closes no nonexistent owner.

- [x] V19 (R3,R22; Cancellation): Hold a read, cancel Program, release read: cancellation arrives before close; late commands refused after admission closure do not access the owner.

- [x] V20 (R3,R19,R22; Outcome): Hold a committed mutation before UI message delivery, cancel Program, then release; runtime receipt reports committed state and readback guidance, never unsaved/retry.

- [x] V21 (R3,R21,R22; Failure): Inject joined run/close errors, pointer/value TransactionError and unknown read cleanup; unknown classification survives cause matching. An earlier successful save followed by a later uncertain or rejected write reports those distinct facts; HadCommittedChanges never hides OutcomeUnknown or labels the rejected write committed. Existing CLI diagnostics remain unchanged.

- [x] V22 (R3,R21; Recovery): Retire old owner before reopening exactly the same absolute database; close failure blocks in-session reopen; failed replacement factory cleans its own partial resources.

- [x] V23 (R22; Signals): Test raw Ctrl+C key, process interrupt and Unix SIGTERM through the existing process signal owner; graceful exit is 1 and terminal restored; ordinary q is 0.

- [x] V24 (R3,R22; Failure): Inject short/error output writer even when the framework ignores returned writer errors; the recording wrapper cancels the run and preserves the failure. Inject command/model panic payload containing private notes; drain resources, report generic failure, and emit no private payload.

- [x] V25 (R3,R6,R22; Boundary): Race queued command start against shutdown admission; no WaitGroup Add-after-Wait race, post-close service call or lost outcome; delayed cleanup keeps owner open and reports progress.

- [x] V26 (R1,R2,R3,R22,R26; Kitty): Owned Kitty with temporary DB: launch, inspect loading shell, q, relaunch, Ctrl+C; verify echo/cursor/alternate screen restored. Record actual visible observation.

### Geometry and terminal presentation — U2

The owner selected spacious two-line task rows in the Kitty prototype on
2026-09-29. Include a terminal-cell SGR regression for the reported shadow
artifact: uniform canvas/dialog backgrounds, consistent whole-backdrop dimming,
and no background bleed at borders or clipping boundaries. Check accented
graphemes and monochrome focus as well as total cell widths. Prototype captures
do not close V39; repeat on the built application. Shared geometry is tested in
U2; later form/consent, saving and Markdown scenarios repeat when U5/U8/U4 add
those states, without claiming they exist in the U2 shell.

- [x] V27 (R9; Layout): At 80x24, 120x40, 200x60 assert outer widths/heights, fixed footer positions and every rendered line's cell width; no trailing newline adds a row.

- [x] V28 (R9,R24; Boundary): Send 0x0, 1x1, 79x24, 80x23, negative synthetic dimensions and repeated grow/shrink; clamp safely, preserve selection/draft and restore layout. Final layout/form matrix passes in `u6-logs/final-resize-cases.log` and `final-checklist-validate.log`.

- [x] V29 (R7,R9; Regression): Switch among empty/loading/error/refreshing/saving and 1/1000 tasks at fixed size; borders and help/status geometry stay unchanged. Final state/profile matrix passes in `u6-logs/final-resize-cases.log` and `final-checklist-validate.log`.

- [x] V30 (R9,R12; Overlay): Center overflowing form/help/consent content within bounds, scroll fields while retaining buttons, and trap focus with Tab/Shift+Tab.

- [x] V31 (R9,R10,R15; Unicode): Render CJK, combining accents, emoji ZWJ clusters, invalid UTF-8 replacement and ten-level indentation at cell boundaries; clip without half-clusters, negative widths or style bleed.

- [x] V32 (R15; Security): ESC, C0/C1, OSC 8/52, bidi controls, CR, DEL and embedded ANSI in every scalar field render as visible safe data, never executable terminal sequences.

- [x] V33 (R15,R23; Security): Multiline display preserves LF, expands tabs to four spaces and escapes remaining controls; raw values remain byte-for-byte unchanged in model/service fixtures.

- [x] V34 (R15; Regression): The shared scalar sanitizer reproduces all existing CLI fixtures and human/JSON separation; extracting it does not change command bytes.

- [x] V35 (R2,R15; Presentation): NO_COLOR strips all content SGR including Markdown; fixed explicit profile avoids appearance queries. Framework cursor/alternate-screen controls are distinguished from content escapes.

- [x] V36 (R12,R15; Accessibility): Every action has a textual key hint and visible focus; status, priority, stale/error and disabled-save meaning survives monochrome output.

- [x] V37 (R5,R9; Purity): After every resize/modal transition, repeated root View calls only return stored frame; no child render, recomposition or I/O occurs.

- [x] V38 (R9,R12,R22; Small terminal): Shrink while editing/saving: draft and operation tokens remain; mutation may finish, Ctrl+C exits, and browse q remains available without triggering form actions.

- [x] V39 (R9,R12,R15,R26; Kitty): Inspect all three supported sizes plus undersized restore, long Unicode and NO_COLOR in owned Kitty; record focus, clipping and visible absence of jitter.

### Projection, keyboard and refresh — U3

- [x] V40 (R10; Normal): Build zero-, one- and ten-level forests; stable root groups and canonical priority/due/created/ID sibling order include done descendants.

- [x] V41 (R10; Boundary): Inject clock around local midnight/DST: Today includes overdue open roots, Upcoming later due roots, Backlog undated, Completed done roots; descendants stay with root.

- [x] V42 (R10,R12; Navigation): j/k/arrows/g/G/pages clamp and scroll; h/l collapse/expand branches and leaf no-op; Tab reverses with Shift+Tab; headers cannot be selected.

- [x] V43 (R10,R24; Selection): External reorder/move preserves selected ID/incarnation; disappearance chooses clamped former index; collapse selects containing ancestor; empty data has no selection.

- [x] V44 (R11; Normal): Literal case-insensitive title/notes substring search follows core semantics, trims search term, finds deep child and preserves only matching/context paths.

- [x] V45 (R11; Boundary): Search that matches a parent excludes unrelated descendants; context nodes are labeled; no matches gives filtered-empty, and clearing restores prefilter collapse state.

- [x] V46 (R8,R11,R24; Debounce): Deliver 150 ms search timer replies out of order, accept with Enter and cancel with Esc; only latest live query applies and Esc restores the entry snapshot.

- [x] V47 (R11; Filters): Status/priority OR, all-of tags and cross-field AND work on the complete forest; empty statuses includes done; Apply/Clear and browse Esc have decided behavior.

- [x] V48 (R11; Dates): Due expressions use injected DayBounds inclusive start/exclusive end, missing dates excluded, invalid expression retained with field error; DST days aren't fixed 24 hours.

- [x] V49 (R11; Dates): Relative due filter resolves on Apply, shows resolved local date and does not drift at midnight; group clock still refreshes.

- [x] V50 (R8; Refresh): Initial/manual/post-write/2-second triggers call service through commands only; active-read triggers coalesce and cannot starve completion with perpetual generation invalidation.

- [x] V51 (R8,R19,R24; Concurrency): Reverse stale forest replies across a mutation/owner change; process matching-owner unknown outcomes before stale-payload rejection, release the busy slot, and prioritize recovery over waiting writes. Known deliberate read cancellation permits the waiting write; other read failure aborts it and preserves draft.

- [x] V52 (R7,R8; Failure): Initial read failure renders retry/quit; refresh failure keeps labeled stale snapshot and disables mutations; successful read clears stale state.

- [x] V53 (R8,R17,R24; Draft): Periodic refresh while editing/help/filter modal stays open cannot steal focus, replace raw fields, reset Base or apply reply to another form instance.

- [x] V54 (R8,R24; Resources): Rapid r/search/selection/ticks produce at most one service operation, one pending refresh/latest history request and bounded timers; no retry storm or accumulating queues.

- [x] V55 (R10,R11,R12; Boundary): Empty/filtered-empty list receives all browse/navigation keys; safe no-ops for absent selection and useful add/clear/search/quit controls.

- [x] V56 (R8,R10,R11,R24,R26; Kitty): Use a second CLI process on isolated DB to edit/reparent/delete during browse and search; inspect live refresh, selection preservation and collapse restoration.

### Details, notes and history — U4

- [x] V57 (R13; Normal): Details show full/wrapped ID, all metadata, optional-value placeholders, numeric progress and timezone/offset; dates remain UTC in stored data.

- [x] V58 (R13,R23; Boundary): Scroll 1 MiB notes, long unbroken text, CJK/combining glyphs and deep tasks at minimum width; no overflow or loss of raw stored notes.

- [x] V59 (R13,R15; Markdown): Render headings, lists, fenced code, tables, inline links and images as terminal text; fixed built-in style/width/profile, no remote image loading or link launch.

- [x] V60 (R13,R15; Security): Malicious Markdown containing OSC/bidi/ANSI/raw HTML, file-like style strings and URL payloads causes no file-style lookup, network, subprocess or terminal-control injection.

- [x] V61 (R13,R24; Concurrency): Delay render then select another task, edit content or resize; only matching ID/incarnation/content/width/profile result enters details.

- [x] V62 (R13; Failure): Renderer error shows safe plain-text fallback; at most one render runs and only newest pending render is retained; no shared mutable renderer.

- [x] V63 (R14; Normal): Timeline displays true ascending int64 sequences, kinds, changed fields and localized time; empty existing history is distinct from missing task.

- [x] V64 (R14,R24; Concurrency): Out-of-order history results for different selections/owner epochs cannot appear under current task; metadata and history freshness are independent.

- [x] V65 (R14; Failure): Known history error is visible and retryable as a read, missing task refreshes forest, unknown cleanup routes to owner recovery; no fabricated events.

- [x] V66 (R13,R14,R19; Integration): After committed edit/status/move, refresh shows actual service metadata events and rollups; history is not claimed atomically consistent with an earlier forest snapshot.

- [x] V67 (R13,R14,R15,R26; Kitty): Inspect Markdown, metadata/history scrolling, no-color text, long notes and selection changes in owned Kitty; safe links remain inert.

### Forms and mutation contracts — U5

- [x] V68 (R12,R16; Create): Create root with minimal title defaults todo/medium/0/empty notes/null parent/due; optional parent chosen by exact ID creates subtask through service.

- [x] V69 (R16; Fields): Title/notes/priority/due/tags/parent and edit-only status/progress map to exact existing command fields; service normalization/date grammar remains authoritative.

- [x] V70 (R16,R23; Patch): Changing one field omits every other field; unchanged raw notes/dates/IDs survive; explicit clear sends ClearDue/ClearParent/empty tags intent.

- [x] V71 (R16,R19; No-op): Empty/equal edit closes with No changes, zero mutation calls and no added history/timestamp update.

- [x] V72 (R12; Focus): Literal q/d/?/x/space in text fields never invokes browse actions; Tab loops form focus, Enter in notes inserts LF, focused Save or Ctrl+S submits.

- [x] V73 (R12,R17; Cancel): Esc on dirty draft defaults Keep editing; explicit Discard closes, clean draft closes directly; restoring focus does not act on a stale selection.

- [x] V74 (R16,R17; Validation): Empty/256-codepoint title, invalid tags/date/status/priority, bad parent, NUL and invalid UTF-8 retain draft and focus first invalid field; no partial mutation.

- [x] V75 (R17; Conflict): Independent owner edits same task after form opens: ErrConflict retains exact Base/draft; Keep draft cannot silently rebase; Reload explicitly discards.

- [x] V76 (R17,R24; Conflict): Independent owner deletes/recreates an ID or changes it then restores equality: obey service incarnation/value semantics; absent task never turns edit into create.

- [x] V77 (R18; Lifecycle): x/Space on open task completes subtree; done task reopens to todo without reopening descendants; repeated satisfied actions create no extra event.

- [x] V78 (R18; Progress): Open leaf accepts 0 and 99, rejects 100/negative/non-numeric; done and parent progress are read-only; parent/status plus dirty progress requires separate saves.

- [x] V79 (R18; Hierarchy): Move to root/other parent recalculates both chains; missing/self/cyclic parent and final depth 11 fail without partial graph/history changes.

- [x] V80 (R18; Policy): With default/true parent-completion policy, add/remove/done/reopen/move verifies direct-child rollup, all-done completion, open-at-100 and last-child removal.

- [x] V81 (R19; Concurrency): Double Enter/Ctrl+S/repeated toggle while read is draining or save pending yields one admitted mutation and one completion; saving blocks Esc/ordinary quit.

- [x] V82 (R17,R19; Failure): Inject known busy/deadline/storage/domain errors; no optimistic persisted values, no retry timer; draft remains available for explicit retry.

- [x] V83 (R19; Partial failure): Commit succeeds, refresh fails: close saved draft, show Saved; refresh failed, block writes until fresh read, never issue second create/edit.

- [x] V84 (R23; Raw preservation): Load controls/tabs and >10,000-line stored notes that a widget cannot round-trip: read-only safe display, omitted patch field, explicit default-Cancel replacement starts blank.

- [x] V85 (R23; Paste): Bracketed paste with LF/Unicode succeeds within supported capacity; invalid/oversized/control input is rejected atomically with previous raw draft intact, not truncated or normalized; OS clipboard command stays disabled.

- [x] V86 (R12,R16,R17,R18,R19,R23,R26; Kitty): Create/edit/move/toggle, multiline paste, validation, conflict, dirty cancel and saved-refresh-failed display in owned Kitty; inspect actual persisted state through CLI JSON.

### Deletion and reconciliation — U8

- [x] V87 (R20; Consent): Leaf d obtains preview, displays title/full ID/count/history loss, initially focuses Cancel; d/y/q and initial Enter cannot delete.

- [x] V88 (R20; Recursive): Parent Delete disabled until explicit subtree checkbox; pass Recursive=true only from checkbox, Expected exact preview and Force=false.

- [x] V89 (R20,R24; Concurrency): Delay preview then close/reopen modal or change selection; mismatched target/incarnation/modal reply cannot arm the current confirmation.

- [x] V90 (R20; Conflict): External target metadata/subtree membership change after preview produces conflict; new preview resets Cancel and recursive checkbox; prior consent cannot carry forward.

- [x] V91 (R20; Failure): Preview failure or removed target writes nothing; known delete failure requires fresh preview for retry; no optimistic deletion from list.

- [x] V92 (R18,R20; Integration): Delete leaf and whole subtree on disk; exact IDs/history removed atomically, ancestor rollup updated, remaining selection chooses correct neighbor.

- [x] V93 (R21; Classification): Value/pointer/joined TransactionError matching canceled/conflict/busy still freezes writes; never route by a retryable cause first.

- [x] V94 (R21; Recovery): Unknown read/write outcome retires old owner, then opens same path for fresh tree/history; old epoch messages cannot unfreeze or overwrite new state. Recovery holds the exclusive service slot without waiting on its own shutdown-drain registration.

- [x] V95 (R21; Failure): Close failure blocks in-session reopen; open/readback failure remains frozen with only read/reopen retry or quit; no DB/WAL/SHM cleanup.

- [x] V96 (R19,R21; Creation): Unknown create with no returned ID retains read-only draft; identical titles/external creations never trigger guessed identity, automatic replay or Retry create.

- [x] V97 (R21; Readback): For old/new state after uncertain edit/delete, present observed state and possible absence; require acknowledgment/discard of uncertain intent before new writes, without claiming causality.

- [x] V98 (R21,R22; Cancellation): Ctrl+C during uncertain write/recovery drains work and preserves unknown diagnostic; preserve unknown outcome and report earlier acknowledged commits separately.

- [x] V99 (R20,R21,R24; Resources): Repeated Reload/confirm keys during active recovery do not overlap owners or writes; departed form/render/history messages remain inert.

- [x] V100 (R20,R21,R26; Kitty): Inspect default Cancel, unchecked recursion, renewed stale consent and quarantined readback in owned Kitty. Label fake-outcome UI injection separately from disk transaction evidence.

### Local acceptance — U6

- [x] V101 (R1,R8,R16,R17,R18,R20,R21,R22,R24; Workflow): Drive load→search→create→edit→external conflict→discard/reload→complete/reopen→move→confirm delete→readback→quit using real disk service plus deterministic message scheduling.

- [x] V102 (R3,R6,R8,R17,R20,R21; Disk/race): Two independent owners exercise WAL reads/writes, cancellation and stale Base/Expected; failure injection proves atomic graph/history and wholly old/new unknown-outcome readback.

- [x] V103 (R3,R22,R26; PTY): Real child PTY tests resize, raw key/paste, Ctrl+C, Unix SIGTERM, failed startup/output and commit-before-cancel; reap children, compare terminal modes/cursor/alternate-screen cleanup.

- [x] V104 (R9,R12,R15,R22,R26; Kitty): Final candidate binary: complete supported-size/undersize/NO_COLOR workflow, interruption/relaunch and external CLI edits in an owned Kitty window with isolated DB and dated observation.

- [x] V105 (R4,R26,R27; Aggregate): Run canonical validate/build/check-generated and Go 1.25 full tests/build-tui on final source; ≥95% each nonexempt package and no generated/module drift.

- [x] V106 (R1,R4,R25; Latency): Run existing three-run CLI benchmark matrix against linked final binary on declared reference host; every case/run meets p90/tail/max limits and retains all samples. Final Konsole-host matrix passes 84/84; see u6-acceptance.json and cli-latency.json.

- [x] V107 (R5,R9,R10,R13,R25; Performance): Measure TUI projection+frame preparation, View, Markdown and first-populated PTY frame separately on declared fixtures; apply defined local preparation budget, retain samples/allocations.

- [x] V108 (R6,R8,R13,R25; Resources): Exercise 10,000-task forest, 1 MiB notes, large history and rapid resize/refresh; record memory/time and verify bounded active/pending commands and no accumulated timers or goroutine growth after settling.

- [x] V109 (R28; Documentation): docs/tui.md/README document command, keys, config, plain/TERM=dumb behavior, unsupported text replacement, save/conflict/delete/unknown outcome and CLI readback; examples match implementation.

- [x] V110 (R26,R27,R28; Review): Review all changed code and current evidence with required lenses; record exact SHA/hashes, unit commits and planning synchronization. No skipped local gate or independent-review claim without evidence. Three completed review receipts, final gate closure, source hashes and unit boundaries are retained in u6-acceptance.json.

### Deferred release proof — Feature 006

- [ ] V111 (R4,R22,R26; Native): Execute exact release candidate on native macOS amd64/arm64 and Windows amd64 plus required Linux targets; prove supported terminal input/resize/cleanup and minimum platform behavior. Cross-builds are insufficient.

- [ ] V112 (R26,R28; Hosted): Feature 006 owns hosted matrix, licenses/release artifacts, installer/removal data preservation, completions/manuals and explicit publication authority; retain link to final native/hosted receipts.

---

## Commands and evidence tiers

Run canonical commands in Codex Bash. Targets below are implemented;
not commands available in the current checkout.

| Tier | Command/method | Owner and evidence |
| --- | --- | --- |
| Focused | `make test-tui TUI_TEST_RUN='<test regex>'` | U1 adds target; U2 includes new terminaltext package; red/green logs per unit |
| CLI regression | `make test-cli CLI_TEST_RUN='<test regex>'` or `make test-cli` | Existing target; tui registration and byte/exit contract receipts |
| Aggregate | `make validate build check-generated` | Existing targets; run on each completed unit's actual contents, retain coverage and race result |
| Minimum Go | `GOTOOLCHAIN=go1.25.0 make test build-tui` | U1 adds build-tui; final full-suite proof in U6; no substitution of ambient compiler |
| Module graph | `make check-modules` | Existing target; retain selected versions and licenses in U1/U6 evidence |
| CLI latency | `make bench-cli CLI_BENCH_OUTPUT=docs/verification-evidence/005/cli-latency.json` | Existing target; final linked binary and all retained samples |
| TUI measurements | `make bench-tui TUI_BENCH_OUTPUT=docs/verification-evidence/005/tui-latency.json` | U6 adds target, formatter/parser tests for sample/report integrity |
| PTY/process | `make test-tui TUI_TEST_RUN='TestTUIProcess|TestSession'` then aggregate full/race | Real child process/PTY evidence; platform skips visible in receipt |
| Visible terminal | Owned Kitty window; matrix below | Unit terminal receipt plus final-candidate observation |
| Native/hosted | Feature 006's future canonical matrix/release commands | V111–V112 remain unchecked here |

The test-tui target invokes existing Go packages with a -run selector through
Make; do not place nonexistent package paths into its U1 version. build-tui
extends the established build-cli pattern to include TUI test compilation and
every later new testable package. No per-package coverage exemption is added.
Integration/PTY cases may skip under -short with explicit reasons; full/race must
execute all host-applicable cases. PTY unsupported-host skips never count as
native execution evidence.

### Performance protocol

Reuse the product/Feature 004 CLI protocol exactly: three complete runs, five
retained warmups and 100 measured fresh-process samples per case/run, nearest-rank
percentiles. Help/version p90/p95/p99/max must be below 5/7.5/10/15 ms; query
limits below 15/20/30/50 ms. Retain target-miss counts, all outliers, output
correctness/size, binary SHA256, compiler, CPU/OS/filesystem, power/governor,
coordinator GOMAXPROCS and source revision. Use a cached binary, current-schema
fixtures and balanced profile, without concurrent tests/builds. No trimming,
subtracting startup cost, best-run selection or repeated sampling until green.
Existing empty/100/1000-task fixtures and all 84 case-runs remain intact.

For TUI, U6 adds three complete fixture runs with five retained warmups and
100 measured samples per case. Measure separately:

- Pure root View returns the stored frame; record time/allocations but do not
  present it as the rendering workload.
- Synchronous projection + layout + final-frame preparation at 80x24, 120x40 and
  200x60 with 0/100/1000 tasks (depth 10, mixed groups and Unicode), a 32 KiB
  selected note and 100 displayed history events. Markdown preprocessing/service
  I/O is excluded and reported separately. On the declared reference host,
  each case/run preparation p95 <16.7 ms, p99 <33.3 ms, max <50 ms. These chosen
  frame-work budgets preserve responsiveness; they are not universal guarantees.
- Markdown cold rendering for 32 KiB/1 MiB notes, service-to-result refresh and
  actual child-PTY startup to first populated frame: retain timings and resource
  observations without inventing a startup SLA. Detect the populated frame via
  fixture content/screen parsing, not a fixed sleep.
- 10,000 tasks, 1 MiB notes, large history, repeated changes and resize storms
  are stress observations. Verify command/timer bounds and eventual quiescence;
  do not silently truncate persisted content or impose unplanned service limits.

Measure frame preparation with actual production code and consume the result.
A fixture budget miss blocks local acceptance and needs a recorded investigation;
do not lower budgets or change fixtures to label it passing. Timing improvements
need a relevant red regression/measurement and canonical verification. Always
retain original failures alongside later evidence.

### Visible Kitty matrix

Use an owned window and a fresh temporary absolute TUSK_DB_PATH; seed via the
built CLI. Record window identity, terminal version/font/profile, binary hash,
screen size, commands/keys, expected/observed output and a screenshot/capture when
available. Observe the actual screen and interactions. A command log or headless
test run inside Kitty is not visual acceptance.

| Stage | Required observation |
| --- | --- |
| U7 | Loading shell, q and Ctrl+C restoration, relaunch, no terminal session reuse |
| U2 | 80x24/120x40/200x60, undersize and restore, Unicode, NO_COLOR, modal bounds |
| U3 | Groups, navigation, collapse, live search/filter, external CLI change |
| U4 | Long notes, Markdown/timeline scrolling, inert links, task switching |
| U5 | Create/edit/move/toggle, paste, validation, conflict, discard, duplicate-submit prevention |
| U8 | Leaf/subtree default-Cancel consent, membership conflict, uncertain readback UI |
| U6 | Repeat complete end-user flow on final binary plus cancellation and clean relaunch |

If desktop/Kitty access is unavailable, record exact blocker and leave its boxes
unchecked. Continue Bash verification. Synthetic frames, PTY emulators, captures
without inspection and cross-builds cannot close these checks. Manual injected
errors are labeled as injected; actual service/SQLite rollback evidence comes
from disk integration tests.

---

## Fixtures and failure injection

- Use deterministic clocks/locations including UTC, America/Sao_Paulo,
  America/New_York DST boundaries, and the service's exceptional civil-day
  fixtures. Use canonical lowercase UUIDv7 fixture IDs and reused opaque IDs
  only for deliberate incarnation tests.
- Build a fake TaskService with per-operation call counters, input clones and
  barrier-controlled replies. Drive typed messages directly; tests never rely
  on real sleeps to order search, selection, render, write or refresh races.
  Executing a returned command is explicit in tests; Update itself must do no I/O.
- Give the session factory close/open counters, failing stages, stable resolved
  path and panic/error payloads. Writer fakes inject zero/partial writes and
  returned failures. Timer/render fakes expose cancellation and generation.
- Deep purity checks start before the first View and retain a structurally
  detached state snapshot; shallow copies or an output-only hash are inadequate.
  Test-only reflection can inspect private widget state without adding production
  reflection/unsafe. A structural guard also prevents root View from delegating
  to child components. Change a nested fixture deliberately to prove the purity
  assertion detects mutation.
- Real temporary disk repositories use WAL and two independent owners. Verify
  mutation plus ancestors/events on a fresh owner, including known rollback and
  wholly old/new uncertain outcomes. Use existing service/storage failure seams;
  do not introduce production SQL corruption switches or relabel fake-service
  outcomes as SQLite proof. Snapshot GetTaskTree and history separately.
- Include 0/1/10 levels, 11-level attempted move, empty/100/1000/10000 tasks,
  equal sort keys, absent/recreated target, large history sequence values,
  zero-event task, long Unicode, control strings, multiline and 1 MiB notes,
  9999/10000/10001-line editor inputs. Verify raw bytes before/after metadata-only
  edits and declined replacement.
- Keyboard tests distinguish bracketed paste from ordinary key messages, literal
  q/d/? in modal input, Enter in notes versus Save, clipboard shortcut disabled,
  saving/shutdown keys, focus restoration and reset consent.
- PTY tests own/reap child processes, use readiness barriers and compare actual
  terminal settings after graceful exit. Keep Ctrl+C and process SIGINT distinct.
  Use Unix-specific fixtures/build constraints; Windows native execution is V111.
- Retained evidence contains fabricated data only. Logs/errors never echo raw
  production titles/notes/paths. Do not delete database sidecars to make a recovery
  test pass.

---

## Execution record

**Planning audit only — 2026-09-29.** Read live source, Makefile, product pack and
versioned upstream source. Audited links, requirement/unit/scenario coverage,
gate ownership and unchecked implementation/release status. No application
tests, dependency build, make validate, benchmark or Kitty check ran.

Implementation receipts belong under `docs/verification-evidence/005/` (future
directory, not fabricated now), one per unit plus final acceptance:

| Field | Required value at execution |
| --- | --- |
| Revision | unit/commit, exact source-tree or source hashes and binary hash |
| Environment | compiler/modules, host/architecture, shell, terminal and fixture |
| Red | exact command, exit/failure and expected defect before implementation |
| Green | focused command/result plus aggregate validation and coverage |
| Scenarios | IDs actually executed, tier, pass/fail/skip and evidence path |
| Terminal | owned window/capture and observed behavior, or pending blocker |
| Findings | workorder IDs fixed/open/deferred and review coverage limits |
| Commit | coherent unit commit before the next unit; no implied publication |

Never backfill a planning checkbox from a later test with different scope.

### U1 local acceptance — 2026-09-29

[Receipt](../verification-evidence/005/u1.json) retains compile and behavioral
Red, focused Green, canonical gates, minimum-Go builds and module inventory.
V06 proves the current widget caches and structural View guarantee; each later
modal/history/form addition must repeat populated purity tests. V10 proves the
cancellable wait primitive; U3 owns bounded timer/debounce chains. No visible
application session exists yet. U7 is the active next unit.

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

### U2 local acceptance — 2026-09-29

[Receipt](../verification-evidence/005/u2.json) retains red/green logs, final
source and binary hashes, canonical validation, minimum-Go race/cross-builds,
and real Kitty screenshots. V27,V31–V34,V37,V39 pass. V28's widget-draft bounds,
V29's current states/task counts, V30's generic overlay/help, V35's current
content, V36's focus and V38's retained widget state pass as partial evidence;
their selection/saving/real form/consent/Markdown portions remain pending in
U3/U4/U5/U8. Keep those boxes open until the stated states actually exist.

The visible resize failure is retained as `u2-kitty/resize-regression-before.png`;
it is failing evidence. The real PTY test was red when the output wrapper hid
`term.File` and green after preserving that interface. Final Kitty screenshots
show resizing and modal restoration without the stale frame or background
bands. No test, build or benchmark ran in Kitty. U3 is next.

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
