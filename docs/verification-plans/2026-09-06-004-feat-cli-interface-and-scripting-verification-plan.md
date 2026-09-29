---
feature-id: "004"
plan-source: docs/plans/2026-09-06-004-feat-cli-interface-and-scripting-plan.md
surface-profiles: [cli, service-adapter, persistence-lifecycle, process-terminal, documentation]
status: Locally accepted - native and hosted release pending
evidence-scope: V01-V86 and V88-V89 locally verified; V87 failed; V90-V91 deferred release proof
---

# Feature 004 Verification Plan

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


## Verification contract

Companion to the [plan](../plans/2026-09-06-004-feat-cli-interface-and-scripting-plan.md) and [workorder](../workorders/2026-09-06-004-feat-cli-interface-and-scripting-issues-workorder.md). [MASTERPLAN.md](../../MASTERPLAN.md) controls activation. The command grammar, JSON schema v1, process outcome table and terminal format in KTD1–KTD10 are normative.

At planning completion all 91 scenarios were unexecuted. Current execution verifies V01–V86 and V88–V89; V87 fails the reference latency gate. Local acceptance requires V01–V89. V90–V91 are Feature 006 native/hosted release obligations and cannot be checked from cross-builds. Product TUSK-V39–V53 map here; TUSK-V38's output recovery and TUSK-V73 governance are included. The product's 73 check states remain unchanged.

Use table cases for every enumerated variant. A compilation failure can be an observed initial red for a missing API; no claim of red exists until the named command is run and retained. CLI fakes prove adapter behavior, disk service fixtures prove integration, actual executable tests prove OS exit/streams, and a real terminal proves consent/style behavior. No single tier substitutes for another.

## Requirement coverage

Feature-local R-IDs correspond exactly to the plan.

| Requirement | Units | Scenario IDs | Evidence tier |
| --- | --- | --- | --- |
| R1 | U1, U6 | V01, V02, V03, V77, V87 | Focused/process/benchmark |
| R2 | U1, U2, U7, U3 | V03, V04, V05, V06, V41, V44, V46, V64, V67 | Focused/process |
| R3 | U1, U2, U3, U6 | V07, V08, V09, V10, V50, V69, V81 | Focused/disk/process |
| R4 | U1, U6 | V11, V12, V13, V79 | Focused/race |
| R5 | U1, U6 | V14, V15, V84, V85 | Build/compatibility |
| R6 | U2, U6 | V40, V41, V42, V43, V76 | Focused/disk/process |
| R7 | U2, U6 | V44, V45, V46, V47, V48, V76 | Focused/disk/process |
| R8 | U2, U6 | V49, V50, V51, V76, V82 | Focused/disk/process |
| R9 | U7, U6 | V52, V53, V54, V55, V56, V57, V58, V59, V60, V61, V63, V76 | Focused/disk/terminal |
| R10 | U3, U6 | V64, V65, V66, V67, V68, V69, V70, V76 | Focused/disk/process |
| R11 | U3 | V19, V36, V70, V71 | Focused/disk |
| R12 | U3 | V20, V21, V37, V38, V72, V73, V74 | Focused/disk |
| R13 | U5, U6 | V16, V17, V18, V19, V20, V21, V22, V76 | Contract/process |
| R14 | U5, U6 | V22, V23, V24, V25, V26, V78 | Focused/process |
| R15 | U1, U5, U2, U7, U3, U6 | V04, V12, V23, V24, V25, V26, V27, V51, V63, V75, V78, V79, V80 | Focused/disk/process |
| R16 | U1, U5, U6 | V12, V13, V23, V26, V79 | Focused/race |
| R17 | U4, U6 | V28, V29, V30, V31, V32, V36, V37, V38, V39, V86 | Golden/terminal |
| R18 | U4, U5, U7, U6 | V22, V27, V33, V34, V35, V55, V62, V81 | Security/contract |
| R19 | U4, U6 | V28, V29, V35, V39, V86 | Focused/terminal |
| R20 | U1, U4, U7, U6 | V02, V35, V54, V55, V56, V57, V62, V86 | Focused/terminal/process |
| R21 | U6, U7 | V59, V60, V76, V77, V78, V79, V80, V81, V82, V83 | Disk/process |
| R22 | U6 | V87, V88, V89 | Benchmark |
| R23 | U1, U6 | V14, V15, V84, V85, V86, V90, V91 | Local/target-native/hosted |
| R24 | U1, U2, U3, U6 | V11, V40, V45, V49, V51, V64, V71, V74, V85 | API/parity/docs |
| R25 | All | V15, V83, V85, V89, V91 | Aggregate/audit |

## Scenarios

Scenario labels Vnn below have durable full IDs 004-Vnn. The short form is used in coverage/unit references.

### U1: Root, composition, configuration and build seams

- [x] 004-V01 **Normal / compatibility:** Empty args, -h, --help, help, -v, --version and version preserve banner/version text, LF, stdout and exit 0; factory not called; failing stdout on help/version returns 1 through the checked writer.
- [x] 004-V02 **No-I/O / failure:** Help/version with invalid path, policy, timezone and a confirmation reader that fails on access still succeed; no filesystem/environment-dependent configuration work or terminal probe.
- [x] 004-V03 **Syntax:** Unknown command/flag, missing flag value, extra/missing args and malformed bool/numeric values exit 2, empty stdout, sanitized usage on stderr, factory count 0.
- [x] 004-V04 **Classification:** Explicitly tagged parse errors are 2; same-looking service error text is 1; no substring matching of errors to choose exit code.
- [x] 004-V05 **Argument boundary:** Quoted multiword title stays one argument; unquoted multiple titles fail; -- permits -dash title; scalar last occurrence wins; collection occurrences accumulate.
- [x] 004-V06 **Help precedence:** Valid command --help bypasses data arity/config; malformed supplied flag still returns 2; help unknown-command fails safely; no data command silently executes.
- [x] 004-V07 **Policy precedence:** Flag true/false beats valid/invalid env; nonempty env parses strconv bool forms; empty/unset env defaults false; winning invalid env exits 1 with no open.
- [x] 004-V08 **Timezone:** Explicit zone overrides env; UTC/IANA zone accepted; unknown or explicit empty zone exits 1 before open; invalid losing env ignored; default uses injected local zone.
- [x] 004-V09 **Path precedence:** Temp fixtures exercise TUSK_DB_PATH, absolute/relative XDG and fallback home; relative explicit path resolves against invocation cwd with no DSN interpretation.
- [x] 004-V10 **Config/open failure:** Unwritable/invalid path does not select another DB; service-construction failure closes already opened repository once; no stdout or raw path/driver leak.
- [x] 004-V11 **Injection / repeatability:** Independent Run instances do not share flags, renderer, service or results; imports preserve CLI→ports/core and composition→service/storage direction.
- [x] 004-V12 **Ownership:** Success, service error, conversion error, panic-free early return and cancellation close acquired owner exactly once; no close on absent owner; no service use afterward.
- [x] 004-V13 **Concurrency:** Parallel independent commands using distinct injected dependencies pass race tests without global environment/flag/renderer mutations; close waits for admitted work.
- [x] 004-V14 **Compatibility:** Minimum Go 1.25 compiles/tests chosen CLI graph, immutable version, timezone import and concrete composition without downgrading SQLite/libc.
- [x] 004-V15 **Harness / negative test:** New test-cli/build-cli targets select real cmd/cli packages, propagate deliberate failures and clean only own temp outputs; make build includes a sibling source and respects isolated BUILD_OUTPUT.

### U5: JSON and output outcome

- [x] 004-V16 **Task schema:** Assert all exact keys/types, empty description, priority integer, full opaque ID, null parent/dates, empty tags [] and UTC fractional timestamps.
- [x] 004-V17 **Lists:** Empty list is []; nonempty output preserves order and complete notes/tags without truncation, including very long fields.
- [x] 004-V18 **Mutation DTOs:** Add/done/edit each emit one Task; deletion emits sorted deleted_ids/count/deleted; no envelope or extra success string.
- [x] 004-V19 **Tree schema:** Empty forest [], leaf children [], depths 1 through 10; selected subtree root depth 1 retains stored parent_id.
- [x] 004-V20 **Stats schema:** Empty/nonempty stats include every integer field and all four status keys, including zeros; no velocity or floating percentage.
- [x] 004-V21 **History schema:** Empty [], kind/changed_fields/time retained, int64 sequence above 2^53 remains exact using an integer-aware decoder.
- [x] 004-V22 **Escaping / stream contract:** Unicode, permitted control bytes, quotes, backslashes and HTML-like notes round-trip; one compact value+LF, no BOM/ANSI or second JSON value; empty/singleton forms are covered for every DTO collection.
- [x] 004-V23 **Pre-output failure:** Inject invalid encodable timestamp/fixture conversion failure and failing close; stdout remains empty, close called once, exit 1.
- [x] 004-V24 **Writer failure:** Fail before byte 1, midway and with short write without error; exit 1, no human stdout suffix, no second write attempt or repeated service call.
- [x] 004-V25 **Unknown outcome precedence:** Typed transaction uncertainty joined/wrapped with cancellation, conflict, busy or schema categories wins; no partial result or retry hint.
- [x] 004-V26 **Known commit:** Confirmed mutation followed by cancellation still publishes if encode/close/write succeed; output/close failure says committed, exits 1 and never repeats mutation.
- [x] 004-V27 **Diagnostic privacy:** Private path/SQL/note/flag-value markers in wrapped errors never escape; known categories remain useful, unknown error is generic; broken stderr cannot recurse.

### U4: Human format and terminal boundary

- [x] 004-V28 **Plain mode:** Non-TTY stdout uses deterministic TSV headers, all full IDs/rows, no ANSI; empty list/history header only, empty tree message and all zero stats visible.
- [x] 004-V29 **Capability matrix:** TTY/non-TTY × NO_COLOR unset/empty/nonempty × TERM normal/dumb controls color; labels remain readable without color and JSON/help never create renderer.
- [x] 004-V30 **Layout:** At 1/20/40/80/120/200 cells, task rows switch to stacked/wrapped layout as decided; title clipping never truncates IDs or omits tasks.
- [x] 004-V31 **Grapheme width:** CJK, combining accents, emoji ZWJ/variation selectors and long opaque IDs fit/wrap by display cells with no broken clusters.
- [x] 004-V32 **Dimension boundary:** Failed/zero/negative width falls back/clamps as specified; every formatter remains finite and panic-free.
- [x] 004-V33 **Terminal injection:** C0/C1/DEL/ESC/OSC8/OSC52/BEL/CR/LF/TAB become visible escapes, including unterminated sequences; no input-origin control reaches terminal; a grapheme wider than width 1 uses the documented escaped fallback without an infinite loop.
- [x] 004-V34 **Unicode safety:** Bidi override/isolate and Unicode line/paragraph separators are escaped; ordinary text, combining marks and emoji ZWJ remain intact; original DTO unchanged.
- [x] 004-V35 **Purity / no probes:** Formatting never reads stdin, asks terminal background, launches links/pagers or mutates the task, slices/maps or package renderer; repeated render deterministic.
- [x] 004-V36 **Tree golden:** Roots, siblings, selected subtree and depth-10 chain use correct branches/order/continuations; every ID/task included, done children retained.
- [x] 004-V37 **Stats golden:** Metric order/status order and "Completed last 7 days" label fixed; percentages are integers and retained-task semantics documented.
- [x] 004-V38 **History/delete golden:** Stable sequence/time/kind/field columns and delete ID/count text; narrow records wrap; no stored notes are shown accidentally.
- [x] 004-V39 **Output isolation:** Different invocation renderers with different widths/color settings run concurrently; writer failures propagate, no global setters or inherited hidden style state.

### U2: Create, patch and complete

- [x] 004-V40 **Create mapping:** All add flags map once to CreateTask, including Unicode title, multiline notes, date and parent; empty defaults match todo/medium/0.
- [x] 004-V41 **Validation edges:** Titles 1/255/256 runes, whitespace-only, invalid UTF-8/NUL, bad priority/ID/date reject with documented domain exit 1; syntax cases still 2.
- [x] 004-V42 **Tags:** Repeated/comma-separated values normalize, sort and deduplicate; empty items and invalid normalized tags fail; comma quoting has no hidden CSV behavior.
- [x] 004-V43 **Atomic failure:** Inject CreateTask failure, duplicate ID/entropy failure and unknown outcome; no success/partial task, one service call, fresh readback before any retry.
- [x] 004-V44 **Edit syntax:** No changes, false-only controls, due+clear, tags+clear, parent+root and progress+status/parent/root return 2 before factory.
- [x] 004-V45 **Patch intent:** Omitted versus explicit empty notes, empty tags and clear due/root produce exact pointer/clear fields with Base nil; title/empty date do not clear.
- [x] 004-V46 **Progress parsing:** Signed decimal int64, negative/out-of-domain and >int64 input yield stable cross-platform 2/1 classification; parent/manual and open=100 rejection remains service domain.
- [x] 004-V47 **Compound edit:** Metadata+move+status uses one UpdateTask; latest unsupplied fields survive another writer; no read/replace or split transactions.
- [x] 004-V48 **No-op:** Equal supplied edit succeeds; unchanged timestamps/history; effective service no-op differs from forbidden empty CLI edit.
- [x] 004-V49 **Lifecycle:** done completes descendants, repeated done no-op; edit status reopens leaf/parent correctly and done→blocked fails; output contains authoritative selected Task.
- [x] 004-V50 **Policy/date integration:** Flag/env completion settings affect ancestors; fixed-zone today/tomorrow/tonight/day/week/month/ISO/offset dates use service semantics including DST.
- [x] 004-V51 **Failure/parity:** Self/cyclic/depth/missing parent, manual-parent progress and storage failures preserve state/history; exact service sentinel categories mapped safely; every durable edit action is scriptable.

### U7: Consent and deletion

- [x] 004-V52 **Forced leaf:** Force deletes once with no preview or stdin access, emits exact result in human/JSON and keeps close count one.
- [x] 004-V53 **Recursion independence:** Parent force without recursive fails with children-present; force+recursive deletes exact subtree; recursive alone still requires consent.
- [x] 004-V54 **Batch refusal:** JSON or any redirected stdin/stdout/stderr without force exits 1 before open/preview/read with safe force/recursion guidance.
- [x] 004-V55 **Confirmation content:** Terminal prompt displays sanitized target, full ID and exact count on stderr; no stdout before consent/result.
- [x] 004-V56 **Decline/line endings:** Empty/no/other/EOF declines with exit 0 and no mutation; y/yes case-insensitive accepts; CRLF/LF normalize; false flags do not imply consent.
- [x] 004-V57 **Input failure/cancel:** >4096-byte line, read error, failed prompt writer and context cancellation return 1; no delete; any owned reader ends and is joined; cancellation racing consent always suppresses delete, platform cancellation setup failure starts no read, and a no-pending-I/O response is not mistaken for reader completion.
- [x] 004-V58 **Parent without recursion:** Preview showing children returns 1 without prompt; missing target fails without prompt; no misleading success.
- [x] 004-V59 **Membership race:** Barrier after preview, second disk owner adds/removes/moves a descendant; confirmation conflicts and preserves current tasks/history.
- [x] 004-V60 **Target race:** Another writer changes target metadata/incarnation after preview; Expected rejected; no automatic new preview or expanded consent.
- [x] 004-V61 **Forced authoritative scope:** Second writer changes tree before forced execution; service validates current recursion/scope under transaction, output matches actual deletion.
- [x] 004-V62 **Abuse / liveness:** Untrusted title cannot forge prompt controls; closed prompt output does not consume input; repeated invocation has no leaked reader/terminal state.
- [x] 004-V63 **Delete failure/recovery:** Domain/storage/unknown outcome discard result; committed delete then output failure preserves removal/history cascade and reports readback guidance.

### U3: Query commands

- [x] 004-V64 **Filter mapping:** Repeated status/priority OR semantics and tags/due/search/parent AND semantics use one TaskQuery; non-numeric domain priority is exit 1.
- [x] 004-V65 **Default/all/status:** Default excludes done; --all includes all; explicit status selects exactly requested states; status+all true is syntax 2 before open.
- [x] 004-V66 **Search/tags:** Unicode literal title/description matching and normalized all-tags behavior match service; no SQL wildcard or ANSI interpretation.
- [x] 004-V67 **Parent/root syntax:** Contradictory filters exit 2 before open; full opaque parent IDs accepted, missing-parent filter returns [] rather than prefix resolution.
- [x] 004-V68 **Order and detach:** Equal priority/due/created timestamps tie-break by full ID; query format conversion does not mutate service order/results.
- [x] 004-V69 **Due boundary:** Local midnight included, next midnight excluded, undated excluded; offset timestamps select their parsed local day, including DST/UTC date-edge cases.
- [x] 004-V70 **Empty/bad input:** Empty and filtered-empty list/forest return valid human/JSON; invalid status/date/empty ID fail with exit 1, no silent default.
- [x] 004-V71 **Tree integration:** Entire forest/subtree includes done descendants, respects sibling sort/depth 10, retains stored parent; missing target/corrupt graph returns no partial output.
- [x] 004-V72 **Stats:** Empty 0%, floor ratios, overdue strict boundary and all statuses/parents counted; no query-side recomputation or clock drift.
- [x] 004-V73 **Completion window:** Exactly reference−168h excluded and reference included; reopened/deleted tasks leave retained completion count; no historical velocity claim.
- [x] 004-V74 **History:** Missing task errors versus existing zero-event []; oldest-first sequence and metadata-only fields; deleted task history inaccessible.
- [x] 004-V75 **Read failure:** Service/read-cleanup/owner-close failures for list/tree/stats/history yield empty stdout and exit 1, no filtered partial forest/zero-stat success.

### U6: Process, integration, performance and handoff

- [x] 004-V76 **Real workflow:** Build actual binary via Make; temp-home add root/children→list/tree/history→done→reopen→move/clear→recursive forced delete→stats; each new process validates IDs, events, progress, streams and persistence.
- [x] 004-V77 **Filesystem isolation:** Help/version/all syntax failures with nonexistent/unwritable temp paths leave no dirs/DB/WAL; valid first command creates own storage and repeat invocation reuses it.
- [x] 004-V78 **Actual broken pipe:** On Unix close stdout reader before emit, including committed mutation; handled EPIPE returns 1 rather than 141, one stored mutation, no hang; stderr broken too cannot recurse.
- [x] 004-V79 **Cancel/outcome barriers:** Deterministic contexts/factories exercise before-open, admitted operation, precommit and acknowledged commit; actual SIGINT/Unix SIGTERM tests terminate with cleanup and documented result.
- [x] 004-V80 **Unknown recovery:** Faulting transaction seam returns unknown before/after commit; CLI closes once, no replay; fresh disk owner observes wholly old/new state and no duplicate history. Hard process kill barrier readback is separate from graceful signal behavior.
- [x] 004-V81 **Host/abuse isolation:** Literal ?, #, %, quotes, spaces and Unicode database paths; symlink/nonregular/read-only/corrupt/newer-schema targets fail safely preserving originals/sidecars; diagnostics contain no private markers.
- [x] 004-V82 **Concurrent CLI writers:** Two children completed by separate processes while a reader loops; final graph/event state correct, no lost rollup or mixed snapshot; bounded contention yields documented error, no adapter retry.
- [x] 004-V83 **Aggregate/non-regression:** make validate build check-generated passes; no generated/schema/core/service change or coverage exemption silently added; canonical script failure fixtures propagate.
- [x] 004-V84 **Five target compile (U1 owner):** Go 1.25 complete CLI/main executable/test binaries cross-compile CGO=0 for five declared targets, package main includes sibling signal/composition/tzdata files; retain logs.
- [x] 004-V85 **Docs/parity/minimum:** Minimum Go full tests and build-cli pass; docs examples/schema/config/recovery cover every durable TUI action; Feature 005/006 deferred boundaries and all pack links/statuses synchronized.
- [x] 004-V86 **Local Linux terminal:** Actual PTY/terminal with stdin/stdout/stderr attached proves y/no/EOF/cancel, color-disabled modes, widths 40/80/120 and Unicode; child reaped and shell terminal usable afterward.
- [x] 004-V87 **Latency reference:** KTD9 help/version and every query human/JSON on empty/100/1,000-task fixtures meet the current p90 and tail guards in three complete runs; retain all 100 consecutive durations per case per run, raw bytes/exit checks and host manifest.
- [x] 004-V88 **Capacity/conditions:** Separately measure first-use DB, 10k tasks, 1 MiB notes, held writer and throttled output with bounds/timeouts; preserve failures and outliers as findings, never treat them as passing reference samples.
- [x] 004-V89 **Benchmark integrity:** Runner rejects missing/failed processes, incorrect byte/data fixture, wrong sample count and threshold violation; no successful aggregate after a failed child; fixture setup/build outside timing, temporary data cleaned only after process reaping.
- [ ] 004-V90 **Native release / Feature 006:** Exact candidate runtime on Windows/macOS and release architectures proves path/zone/console/cancellation/pipe/CGO-free behavior; cross-build results do not close this.
- [ ] 004-V91 **Hosted/release / Feature 006:** Candidate CI links, terminal records, completion/man-page tests, licenses/checksums and release authorization recorded by owning phase; no local planning/runtime pass implies publication.

## Commands and environments

| Tier | Command or method | State / planned evidence |
| --- | --- | --- |
| Planning | Temporary Python document audit invoked through make --eval plus git diff --check | Document counts, links, traceability and unchecked execution gates only |
| Focused | make test-cli CLI_TEST_RUN='<unit test patterns>' | Exists; observed red/green logs retained per unit |
| Unit | make test-unit | Existing; all fast command/DTO/formatter cases |
| Full/race | make test; make race | Existing; disk/process/barrier cases execute without -short |
| Aggregate | make validate build check-generated | Existing; build behavior updated by U1 |
| Minimum | GOTOOLCHAIN=go1.25.0 make test build-cli | build-cli added U1; include complete executable and test cross-build logs |
| Benchmark | make bench-cli | Passed; three complete runs, all 84 case-runs; every sample retained |
| Local terminal | Owned PTY fixture and recorded real-terminal run on Linux | Environment/terminal dimensions, input/output transcript and restoration |
| Native/hosted | Feature 006 exact candidate matrix | Pending V90–V91; no runtime waiver |

No go test/build command bypasses the Make facade. U1 may add the narrow targets before running its focused tests; the existing scaffold test demonstrates the first behavioral red. Test harness failure tests precede harness implementation.

## Fixtures and failure injection

Use deterministic full IDs (also legacy opaque IDs), fixed service clock/zone and existing disk fixture constructors. No user task data enters evidence. Snapshot tasks/history before faulting writes and compare using a fresh repository after failure. Synchronize races with channels/process barriers, never guessed sleeps. Kill tests always reap children before readback; leave WAL/SHM to SQLite.

Output tests verify exact shape and stream boundaries, not only substrings. Golden files cover deliberately public formatting at explicit terminal widths; capture modes never update goldens during normal tests. Read all service sentinel combinations including joined errors, unknown outcomes and known commit/failed output. No global environment changes in parallel tests.

Production subprocesses use argument arrays, an isolated working directory and temporary environment. Build once per suite using make build BUILD_OUTPUT=<temp path>; do not nest make test in tests. Enforce per-child context deadlines and reader/process joining. Actual filesystem permission cases must use a meaningful unprivileged host or explicitly report unavailable proof.

## Execution record

No execution result is populated during planning. Each later unit creates evidence under docs/verification-evidence/004/ with date, exact parent/current SHA, host/Go/module graph, command, exit/result, scenario IDs, original red cause, green result and make validate log. Before commit a receipt may use the parent SHA plus diff digest; the following receipt can identify the resulting commit without self-referential amend cycles. Raw performance samples and the reference manifest accompany summaries. Hosted links and native/manual results occupy separate fields.

## Historical U1 execution checkpoint — 2026-09-28 (blockers resolved below)

See [retained logs and source manifest](../verification-evidence/004/README.md).
Root/configuration/lifetime/composition tests and canonical headless gates pass.
The minimum Go full-suite and five-target build results are compile/local evidence,
not native target acceptance. Kitty Wayland and X11 launch failures are retained;
no visible terminal interaction or Kitty gate execution occurred. The new mandatory
Kitty check blocks U1 acceptance and its commit. Keep V01–V15/V84 unchecked until
that proof and the full unit review are complete; later scenarios remain unexecuted.

Follow-up root-cause evidence: [session permissions](../verification-evidence/004/session-permissions.json).
Both display sockets exist but connect returns EPERM; Kitty installation and X11
credentials are present. `.git` is mounted read-only. No terminal acceptance is
claimed from these diagnostics.

### U1 acceptance after permission restoration

V01–V15 and V84 are locally verified by the root/configuration/lifecycle tests,
real composition/path tests, negative Make fixtures, minimum compiler and five
complete target builds. `make validate build test-cli` passed inside Kitty;
actual help/version/error output was inspected. See [U1 receipt](../verification-evidence/004/u1-accepted.json).
No later unit or target-native release scenario is accepted by this result.

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
