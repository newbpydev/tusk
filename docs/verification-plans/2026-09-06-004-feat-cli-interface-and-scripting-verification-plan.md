---
feature-id: "004"
plan-source: docs/plans/2026-09-06-004-feat-cli-interface-and-scripting-plan.md
surface-profiles: [cli, service-adapter, persistence-lifecycle, process-terminal, documentation]
status: Planned - not executed
evidence-scope: Planning only
---

# Feature 004 Verification Plan

## Verification contract

Companion to the [plan](../plans/2026-09-06-004-feat-cli-interface-and-scripting-plan.md) and [workorder](../workorders/2026-09-06-004-feat-cli-interface-and-scripting-issues-workorder.md). [MASTERPLAN.md](../../MASTERPLAN.md) controls activation. The command grammar, JSON schema v1, process outcome table and terminal format in KTD1–KTD10 are normative.

All 91 scenarios are planned, not executed. Local acceptance requires V01–V89. V90–V91 are Feature 006 native/hosted release obligations and cannot be checked from cross-builds. Product TUSK-V39–V53 map here; TUSK-V38's output recovery and TUSK-V73 governance are included. The product's 73 check states remain unchanged.

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

- [ ] 004-V01 **Normal / compatibility:** Empty args, -h, --help, help, -v, --version and version preserve banner/version text, LF, stdout and exit 0; factory not called; failing stdout on help/version returns 1 through the checked writer.
- [ ] 004-V02 **No-I/O / failure:** Help/version with invalid path, policy, timezone and a confirmation reader that fails on access still succeed; no filesystem/environment-dependent configuration work or terminal probe.
- [ ] 004-V03 **Syntax:** Unknown command/flag, missing flag value, extra/missing args and malformed bool/numeric values exit 2, empty stdout, sanitized usage on stderr, factory count 0.
- [ ] 004-V04 **Classification:** Explicitly tagged parse errors are 2; same-looking service error text is 1; no substring matching of errors to choose exit code.
- [ ] 004-V05 **Argument boundary:** Quoted multiword title stays one argument; unquoted multiple titles fail; -- permits -dash title; scalar last occurrence wins; collection occurrences accumulate.
- [ ] 004-V06 **Help precedence:** Valid command --help bypasses data arity/config; malformed supplied flag still returns 2; help unknown-command fails safely; no data command silently executes.
- [ ] 004-V07 **Policy precedence:** Flag true/false beats valid/invalid env; nonempty env parses strconv bool forms; empty/unset env defaults false; winning invalid env exits 1 with no open.
- [ ] 004-V08 **Timezone:** Explicit zone overrides env; UTC/IANA zone accepted; unknown or explicit empty zone exits 1 before open; invalid losing env ignored; default uses injected local zone.
- [ ] 004-V09 **Path precedence:** Temp fixtures exercise TUSK_DB_PATH, absolute/relative XDG and fallback home; relative explicit path resolves against invocation cwd with no DSN interpretation.
- [ ] 004-V10 **Config/open failure:** Unwritable/invalid path does not select another DB; service-construction failure closes already opened repository once; no stdout or raw path/driver leak.
- [ ] 004-V11 **Injection / repeatability:** Independent Run instances do not share flags, renderer, service or results; imports preserve CLI→ports/core and composition→service/storage direction.
- [ ] 004-V12 **Ownership:** Success, service error, conversion error, panic-free early return and cancellation close acquired owner exactly once; no close on absent owner; no service use afterward.
- [ ] 004-V13 **Concurrency:** Parallel independent commands using distinct injected dependencies pass race tests without global environment/flag/renderer mutations; close waits for admitted work.
- [ ] 004-V14 **Compatibility:** Minimum Go 1.25 compiles/tests chosen CLI graph, immutable version, timezone import and concrete composition without downgrading SQLite/libc.
- [ ] 004-V15 **Harness / negative test:** New test-cli/build-cli targets select real cmd/cli packages, propagate deliberate failures and clean only own temp outputs; make build includes a sibling source and respects isolated BUILD_OUTPUT.

### U5: JSON and output outcome

- [ ] 004-V16 **Task schema:** Assert all exact keys/types, empty description, priority integer, full opaque ID, null parent/dates, empty tags [] and UTC fractional timestamps.
- [ ] 004-V17 **Lists:** Empty list is []; nonempty output preserves order and complete notes/tags without truncation, including very long fields.
- [ ] 004-V18 **Mutation DTOs:** Add/done/edit each emit one Task; deletion emits sorted deleted_ids/count/deleted; no envelope or extra success string.
- [ ] 004-V19 **Tree schema:** Empty forest [], leaf children [], depths 1 through 10; selected subtree root depth 1 retains stored parent_id.
- [ ] 004-V20 **Stats schema:** Empty/nonempty stats include every integer field and all four status keys, including zeros; no velocity or floating percentage.
- [ ] 004-V21 **History schema:** Empty [], kind/changed_fields/time retained, int64 sequence above 2^53 remains exact using an integer-aware decoder.
- [ ] 004-V22 **Escaping / stream contract:** Unicode, permitted control bytes, quotes, backslashes and HTML-like notes round-trip; one compact value+LF, no BOM/ANSI or second JSON value; empty/singleton forms are covered for every DTO collection.
- [ ] 004-V23 **Pre-output failure:** Inject invalid encodable timestamp/fixture conversion failure and failing close; stdout remains empty, close called once, exit 1.
- [ ] 004-V24 **Writer failure:** Fail before byte 1, midway and with short write without error; exit 1, no human stdout suffix, no second write attempt or repeated service call.
- [ ] 004-V25 **Unknown outcome precedence:** Typed transaction uncertainty joined/wrapped with cancellation, conflict, busy or schema categories wins; no partial result or retry hint.
- [ ] 004-V26 **Known commit:** Confirmed mutation followed by cancellation still publishes if encode/close/write succeed; output/close failure says committed, exits 1 and never repeats mutation.
- [ ] 004-V27 **Diagnostic privacy:** Private path/SQL/note/flag-value markers in wrapped errors never escape; known categories remain useful, unknown error is generic; broken stderr cannot recurse.

### U4: Human format and terminal boundary

- [ ] 004-V28 **Plain mode:** Non-TTY stdout uses deterministic TSV headers, all full IDs/rows, no ANSI; empty list/history header only, empty tree message and all zero stats visible.
- [ ] 004-V29 **Capability matrix:** TTY/non-TTY × NO_COLOR unset/empty/nonempty × TERM normal/dumb controls color; labels remain readable without color and JSON/help never create renderer.
- [ ] 004-V30 **Layout:** At 1/20/40/80/120/200 cells, task rows switch to stacked/wrapped layout as decided; title clipping never truncates IDs or omits tasks.
- [ ] 004-V31 **Grapheme width:** CJK, combining accents, emoji ZWJ/variation selectors and long opaque IDs fit/wrap by display cells with no broken clusters.
- [ ] 004-V32 **Dimension boundary:** Failed/zero/negative width falls back/clamps as specified; every formatter remains finite and panic-free.
- [ ] 004-V33 **Terminal injection:** C0/C1/DEL/ESC/OSC8/OSC52/BEL/CR/LF/TAB become visible escapes, including unterminated sequences; no input-origin control reaches terminal; a grapheme wider than width 1 uses the documented escaped fallback without an infinite loop.
- [ ] 004-V34 **Unicode safety:** Bidi override/isolate and Unicode line/paragraph separators are escaped; ordinary text, combining marks and emoji ZWJ remain intact; original DTO unchanged.
- [ ] 004-V35 **Purity / no probes:** Formatting never reads stdin, asks terminal background, launches links/pagers or mutates the task, slices/maps or package renderer; repeated render deterministic.
- [ ] 004-V36 **Tree golden:** Roots, siblings, selected subtree and depth-10 chain use correct branches/order/continuations; every ID/task included, done children retained.
- [ ] 004-V37 **Stats golden:** Metric order/status order and "Completed last 7 days" label fixed; percentages are integers and retained-task semantics documented.
- [ ] 004-V38 **History/delete golden:** Stable sequence/time/kind/field columns and delete ID/count text; narrow records wrap; no stored notes are shown accidentally.
- [ ] 004-V39 **Output isolation:** Different invocation renderers with different widths/color settings run concurrently; writer failures propagate, no global setters or inherited hidden style state.

### U2: Create, patch and complete

- [ ] 004-V40 **Create mapping:** All add flags map once to CreateTask, including Unicode title, multiline notes, date and parent; empty defaults match todo/medium/0.
- [ ] 004-V41 **Validation edges:** Titles 1/255/256 runes, whitespace-only, invalid UTF-8/NUL, bad priority/ID/date reject with documented domain exit 1; syntax cases still 2.
- [ ] 004-V42 **Tags:** Repeated/comma-separated values normalize, sort and deduplicate; empty items and invalid normalized tags fail; comma quoting has no hidden CSV behavior.
- [ ] 004-V43 **Atomic failure:** Inject CreateTask failure, duplicate ID/entropy failure and unknown outcome; no success/partial task, one service call, fresh readback before any retry.
- [ ] 004-V44 **Edit syntax:** No changes, false-only controls, due+clear, tags+clear, parent+root and progress+status/parent/root return 2 before factory.
- [ ] 004-V45 **Patch intent:** Omitted versus explicit empty notes, empty tags and clear due/root produce exact pointer/clear fields with Base nil; title/empty date do not clear.
- [ ] 004-V46 **Progress parsing:** Signed decimal int64, negative/out-of-domain and >int64 input yield stable cross-platform 2/1 classification; parent/manual and open=100 rejection remains service domain.
- [ ] 004-V47 **Compound edit:** Metadata+move+status uses one UpdateTask; latest unsupplied fields survive another writer; no read/replace or split transactions.
- [ ] 004-V48 **No-op:** Equal supplied edit succeeds; unchanged timestamps/history; effective service no-op differs from forbidden empty CLI edit.
- [ ] 004-V49 **Lifecycle:** done completes descendants, repeated done no-op; edit status reopens leaf/parent correctly and done→blocked fails; output contains authoritative selected Task.
- [ ] 004-V50 **Policy/date integration:** Flag/env completion settings affect ancestors; fixed-zone today/tomorrow/tonight/day/week/month/ISO/offset dates use service semantics including DST.
- [ ] 004-V51 **Failure/parity:** Self/cyclic/depth/missing parent, manual-parent progress and storage failures preserve state/history; exact service sentinel categories mapped safely; every durable edit action is scriptable.

### U7: Consent and deletion

- [ ] 004-V52 **Forced leaf:** Force deletes once with no preview or stdin access, emits exact result in human/JSON and keeps close count one.
- [ ] 004-V53 **Recursion independence:** Parent force without recursive fails with children-present; force+recursive deletes exact subtree; recursive alone still requires consent.
- [ ] 004-V54 **Batch refusal:** JSON or any redirected stdin/stdout/stderr without force exits 1 before open/preview/read with safe force/recursion guidance.
- [ ] 004-V55 **Confirmation content:** Terminal prompt displays sanitized target, full ID and exact count on stderr; no stdout before consent/result.
- [ ] 004-V56 **Decline/line endings:** Empty/no/other/EOF declines with exit 0 and no mutation; y/yes case-insensitive accepts; CRLF/LF normalize; false flags do not imply consent.
- [ ] 004-V57 **Input failure/cancel:** >4096-byte line, read error, failed prompt writer and context cancellation return 1; no delete; any owned reader ends and is joined; cancellation racing consent always suppresses delete, platform cancellation setup failure starts no read, and a no-pending-I/O response is not mistaken for reader completion.
- [ ] 004-V58 **Parent without recursion:** Preview showing children returns 1 without prompt; missing target fails without prompt; no misleading success.
- [ ] 004-V59 **Membership race:** Barrier after preview, second disk owner adds/removes/moves a descendant; confirmation conflicts and preserves current tasks/history.
- [ ] 004-V60 **Target race:** Another writer changes target metadata/incarnation after preview; Expected rejected; no automatic new preview or expanded consent.
- [ ] 004-V61 **Forced authoritative scope:** Second writer changes tree before forced execution; service validates current recursion/scope under transaction, output matches actual deletion.
- [ ] 004-V62 **Abuse / liveness:** Untrusted title cannot forge prompt controls; closed prompt output does not consume input; repeated invocation has no leaked reader/terminal state.
- [ ] 004-V63 **Delete failure/recovery:** Domain/storage/unknown outcome discard result; committed delete then output failure preserves removal/history cascade and reports readback guidance.

### U3: Query commands

- [ ] 004-V64 **Filter mapping:** Repeated status/priority OR semantics and tags/due/search/parent AND semantics use one TaskQuery; non-numeric domain priority is exit 1.
- [ ] 004-V65 **Default/all/status:** Default excludes done; --all includes all; explicit status selects exactly requested states; status+all true is syntax 2 before open.
- [ ] 004-V66 **Search/tags:** Unicode literal title/description matching and normalized all-tags behavior match service; no SQL wildcard or ANSI interpretation.
- [ ] 004-V67 **Parent/root syntax:** Contradictory filters exit 2 before open; full opaque parent IDs accepted, missing-parent filter returns [] rather than prefix resolution.
- [ ] 004-V68 **Order and detach:** Equal priority/due/created timestamps tie-break by full ID; query format conversion does not mutate service order/results.
- [ ] 004-V69 **Due boundary:** Local midnight included, next midnight excluded, undated excluded; offset timestamps select their parsed local day, including DST/UTC date-edge cases.
- [ ] 004-V70 **Empty/bad input:** Empty and filtered-empty list/forest return valid human/JSON; invalid status/date/empty ID fail with exit 1, no silent default.
- [ ] 004-V71 **Tree integration:** Entire forest/subtree includes done descendants, respects sibling sort/depth 10, retains stored parent; missing target/corrupt graph returns no partial output.
- [ ] 004-V72 **Stats:** Empty 0%, floor ratios, overdue strict boundary and all statuses/parents counted; no query-side recomputation or clock drift.
- [ ] 004-V73 **Completion window:** Exactly reference−168h excluded and reference included; reopened/deleted tasks leave retained completion count; no historical velocity claim.
- [ ] 004-V74 **History:** Missing task errors versus existing zero-event []; oldest-first sequence and metadata-only fields; deleted task history inaccessible.
- [ ] 004-V75 **Read failure:** Service/read-cleanup/owner-close failures for list/tree/stats/history yield empty stdout and exit 1, no filtered partial forest/zero-stat success.

### U6: Process, integration, performance and handoff

- [ ] 004-V76 **Real workflow:** Build actual binary via Make; temp-home add root/children→list/tree/history→done→reopen→move/clear→recursive forced delete→stats; each new process validates IDs, events, progress, streams and persistence.
- [ ] 004-V77 **Filesystem isolation:** Help/version/all syntax failures with nonexistent/unwritable temp paths leave no dirs/DB/WAL; valid first command creates own storage and repeat invocation reuses it.
- [ ] 004-V78 **Actual broken pipe:** On Unix close stdout reader before emit, including committed mutation; handled EPIPE returns 1 rather than 141, one stored mutation, no hang; stderr broken too cannot recurse.
- [ ] 004-V79 **Cancel/outcome barriers:** Deterministic contexts/factories exercise before-open, admitted operation, precommit and acknowledged commit; actual SIGINT/Unix SIGTERM tests terminate with cleanup and documented result.
- [ ] 004-V80 **Unknown recovery:** Faulting transaction seam returns unknown before/after commit; CLI closes once, no replay; fresh disk owner observes wholly old/new state and no duplicate history. Hard process kill barrier readback is separate from graceful signal behavior.
- [ ] 004-V81 **Host/abuse isolation:** Literal ?, #, %, quotes, spaces and Unicode database paths; symlink/nonregular/read-only/corrupt/newer-schema targets fail safely preserving originals/sidecars; diagnostics contain no private markers.
- [ ] 004-V82 **Concurrent CLI writers:** Two children completed by separate processes while a reader loops; final graph/event state correct, no lost rollup or mixed snapshot; bounded contention yields documented error, no adapter retry.
- [ ] 004-V83 **Aggregate/non-regression:** make validate build check-generated passes; no generated/schema/core/service change or coverage exemption silently added; canonical script failure fixtures propagate.
- [ ] 004-V84 **Five target compile (U1 owner):** Go 1.25 complete CLI/main executable/test binaries cross-compile CGO=0 for five declared targets, package main includes sibling signal/composition/tzdata files; retain logs.
- [ ] 004-V85 **Docs/parity/minimum:** Minimum Go full tests and build-cli pass; docs examples/schema/config/recovery cover every durable TUI action; Feature 005/006 deferred boundaries and all pack links/statuses synchronized.
- [ ] 004-V86 **Local Linux terminal:** Actual PTY/terminal with stdin/stdout/stderr attached proves y/no/EOF/cancel, color-disabled modes, widths 40/80/120 and Unicode; child reaped and shell terminal usable afterward.
- [ ] 004-V87 **Latency reference:** KTD9 help/version and every query human/JSON on empty/100/1,000-task fixtures meet every-sample <5/<15 ms; retain all 100 consecutive durations per case, raw bytes/exit checks and host manifest.
- [ ] 004-V88 **Capacity/conditions:** Separately measure first-use DB, 10k tasks, 1 MiB notes, held writer and throttled output with bounds/timeouts; preserve failures and outliers as findings, never treat them as passing reference samples.
- [ ] 004-V89 **Benchmark integrity:** Runner rejects missing/failed processes, incorrect byte/data fixture, wrong sample count and threshold violation; no successful aggregate after a failed child; fixture setup/build outside timing, temporary data cleaned only after process reaping.
- [ ] 004-V90 **Native release / Feature 006:** Exact candidate runtime on Windows/macOS and release architectures proves path/zone/console/cancellation/pipe/CGO-free behavior; cross-build results do not close this.
- [ ] 004-V91 **Hosted/release / Feature 006:** Candidate CI links, terminal records, completion/man-page tests, licenses/checksums and release authorization recorded by owning phase; no local planning/runtime pass implies publication.

## Commands and environments

| Tier | Command or method | State / planned evidence |
| --- | --- | --- |
| Planning | Temporary Python document audit invoked through make --eval plus git diff --check | Document counts, links, traceability and unchecked execution gates only |
| Focused | make test-cli CLI_TEST_RUN='<unit test patterns>' | Planned U1 target; retain exact observed red then green |
| Unit | make test-unit | Existing; all fast command/DTO/formatter cases |
| Full/race | make test; make race | Existing; disk/process/barrier cases execute without -short |
| Aggregate | make validate build check-generated | Existing; build behavior updated by U1 |
| Minimum | GOTOOLCHAIN=go1.25.0 make test build-cli | build-cli added U1; include complete executable and test cross-build logs |
| Benchmark | make bench-cli | Planned U6 target; actual production binary and KTD9 fixed fixtures |
| Local terminal | Owned PTY fixture and recorded real-terminal run on Linux | Environment/terminal dimensions, input/output transcript and restoration |
| Native/hosted | Feature 006 exact candidate matrix | Pending V90–V91; no runtime waiver |

No go test/build command bypasses the Make facade. U1 may add the narrow targets before running its focused tests; the existing scaffold test demonstrates the first behavioral red. Test harness failure tests precede harness implementation.

## Fixtures and failure injection

Use deterministic full IDs (also legacy opaque IDs), fixed service clock/zone and existing disk fixture constructors. No user task data enters evidence. Snapshot tasks/history before faulting writes and compare using a fresh repository after failure. Synchronize races with channels/process barriers, never guessed sleeps. Kill tests always reap children before readback; leave WAL/SHM to SQLite.

Output tests verify exact shape and stream boundaries, not only substrings. Golden files cover deliberately public formatting at explicit terminal widths; capture modes never update goldens during normal tests. Read all service sentinel combinations including joined errors, unknown outcomes and known commit/failed output. No global environment changes in parallel tests.

Production subprocesses use argument arrays, an isolated working directory and temporary environment. Build once per suite using make build BUILD_OUTPUT=<temp path>; do not nest make test in tests. Enforce per-child context deadlines and reader/process joining. Actual filesystem permission cases must use a meaningful unprivileged host or explicitly report unavailable proof.

## Execution record

No execution result is populated during planning. Each later unit creates evidence under docs/verification-evidence/004/ with date, exact parent/current SHA, host/Go/module graph, command, exit/result, scenario IDs, original red cause, green result and make validate log. Before commit a receipt may use the parent SHA plus diff digest; the following receipt can identify the resulting commit without self-referential amend cycles. Raw performance samples and the reference manifest accompany summaries. Hosted links and native/manual results occupy separate fields.
