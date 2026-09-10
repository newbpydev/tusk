---
feature-id: TUSK
plan-source: docs/plans/2026-09-06-001-feat-tusk-modern-task-system-plan.md
surface-profiles: [cli-tui, library-service, persistence-migration, packaging-operations, documentation]
status: Planned - not executed
evidence-scope: Planning only
---

# Tusk Modern Task System Verification Plan

## Verification contract

This is the product-level companion to the [plan](../plans/2026-09-06-001-feat-tusk-modern-task-system-plan.md) and [workorder](../workorders/2026-09-06-001-feat-tusk-modern-task-system-issues-workorder.md). [MASTERPLAN.md](../../MASTERPLAN.md) controls phase activation. Each feature owner must import the applicable scenarios into that feature's synchronized triplet before execution (G1).

Behavior under test: R1–R29, F1–F4 and the CLI wire/key/mutation contracts. Existing core evidence is historical; all scenarios in this document are unexecuted. Public machines consume CLI DTOs, not internal struct tags. Native Linux/macOS/Windows and the five release targets require separate evidence. No browser/device scenarios are needed for a terminal application.

Local tiers are focused assertions within canonical suites, aggregate quality gates, disk/process integration, and benchmarks. Hosted tiers prove exact-revision OS/architecture behavior. Manual tiers prove actual terminal/lifecycle usability. Module/document inspection cannot close runtime gates. Feature 002 records local acceptance of G2's pins and six units; product U24 / feature U6 supplied its compatibility prerequisite. Native/hosted acceptance remains Phase 6, and Feature 003 planning reran no runtime tests. The product pack now has 24 handoff units and retains all 73 scenario IDs.

## Requirement coverage

| Requirement | Units | Scenario IDs | Evidence tier |
| --- | --- | --- | --- |
| R1 | U3, U11, U15, U21, U23 | TUSK-V08, TUSK-V39, TUSK-V40, TUSK-V50, TUSK-V51, TUSK-V67, TUSK-V72 | Local/process/hosted |
| R2 | U3, U11, U15, U22 | TUSK-V08, TUSK-V09, TUSK-V41, TUSK-V53, TUSK-V70 | Local/process/hosted |
| R3 | U1, U4, U6, U7, U12 | TUSK-V12, TUSK-V20, TUSK-V25, TUSK-V42 | Local/integration |
| R4 | U1, U4, U6, U12, U19 | TUSK-V12, TUSK-V13, TUSK-V20, TUSK-V42, TUSK-V61 | Local/integration |
| R5 | U6, U8, U12, U15 | TUSK-V21, TUSK-V28, TUSK-V34, TUSK-V42, TUSK-V50 | Local/process |
| R6 | U4, U8, U10 | TUSK-V06, TUSK-V15, TUSK-V27, TUSK-V31, TUSK-V35 | Local/disk/concurrency |
| R7 | U4, U8, U18 | TUSK-V26, TUSK-V28, TUSK-V30, TUSK-V59 | Local/integration |
| R8 | U8, U10, U12, U19 | TUSK-V28, TUSK-V31, TUSK-V42, TUSK-V64 | Local/process/UI |
| R9 | U6, U8, U11 | TUSK-V29, TUSK-V41, TUSK-V64 | Local/process/UI |
| R10 | U1, U4, U5, U8, U10 | TUSK-V14, TUSK-V16, TUSK-V18, TUSK-V31, TUSK-V38 | Disk/concurrency/process |
| R11 | U1, U3, U22 | TUSK-V02, TUSK-V03, TUSK-V04, TUSK-V70 | Disk/recovery/hosted |
| R12 | U24, U3, U5, U21 | TUSK-V01, TUSK-V10, TUSK-V11, TUSK-V16, TUSK-V17, TUSK-V18, TUSK-V19, TUSK-V67 | Disk/process/hosted |
| R13 | U2, U4, U9, U13, U17 | TUSK-V07, TUSK-V32, TUSK-V46, TUSK-V56 | Local/integration/UI |
| R14 | U2, U9, U14, U17 | TUSK-V07, TUSK-V32, TUSK-V46, TUSK-V56 | Local/process/UI |
| R15 | U7, U9, U18 | TUSK-V22, TUSK-V23, TUSK-V24, TUSK-V59 | Local/portability |
| R16 | U7, U9, U13 | TUSK-V24, TUSK-V32, TUSK-V46 | Local/process |
| R17 | U9, U13 | TUSK-V33, TUSK-V45, TUSK-V47 | Local/process |
| R18 | U1, U2, U4, U8, U9, U13, U18 | TUSK-V14, TUSK-V28, TUSK-V34, TUSK-V38, TUSK-V45, TUSK-V59 | Disk/process/UI |
| R19 | U6, U10, U12, U13, U19 | TUSK-V20, TUSK-V21, TUSK-V36, TUSK-V42, TUSK-V44, TUSK-V61, TUSK-V63 | Local/process/UI |
| R20 | U8, U10, U12, U19 | TUSK-V30, TUSK-V37, TUSK-V43, TUSK-V61, TUSK-V63 | Concurrency/process/UI |
| R21 | U12, U13, U14, U15 | TUSK-V43, TUSK-V45, TUSK-V47, TUSK-V49, TUSK-V50 | Local/process |
| R22 | U6, U10, U11, U14, U15, U19 | TUSK-V17, TUSK-V38, TUSK-V40, TUSK-V49, TUSK-V52, TUSK-V62 | Local/process/UI |
| R23 | U3, U11, U14, U18, U20 | TUSK-V09, TUSK-V48, TUSK-V53, TUSK-V60, TUSK-V66 | Security/process/manual |
| R24 | U16, U17, U18, U19, U20 | TUSK-V54, TUSK-V57, TUSK-V58, TUSK-V64, TUSK-V65 | Synthetic/race/benchmark |
| R25 | U17, U19, U20 | TUSK-V57, TUSK-V58, TUSK-V62, TUSK-V63, TUSK-V64 | Synthetic/integration/manual |
| R26 | U16, U17, U18, U19, U20 | TUSK-V55, TUSK-V56, TUSK-V59, TUSK-V61, TUSK-V66 | Synthetic/manual |
| R27 | U24, U21, U22, U23 | TUSK-V01, TUSK-V67, TUSK-V68, TUSK-V69, TUSK-V70, TUSK-V71, TUSK-V72 | Local cross-build/hosted/manual/release |
| R28 | U5, U15, U20 | TUSK-V19, TUSK-V51, TUSK-V65 | Benchmark |
| R29 | All units | TUSK-V01, TUSK-V05, TUSK-V67, TUSK-V73 | Local/hosted/documentation |

## Scenarios

Each scenario includes its fixture/action and expected outcome. The unit's named test/owned test paths locate the executable assertions. Test names and failures are planned, not observed. Broad scenarios must be split into table cases at implementation; no single happy-path E2E test can replace them.

### U24. Pinned storage runtime and compatibility proof

- [ ] TUSK-V01 **Compatibility / runtime:** Feature 002 U6 proves Go 1.25 minimum, modernc v1.58.0, libc v1.75.6 and SQLite 3.53.4 through named compatibility tests and five CGO-disabled target builds. Native runtime engine proof on each target remains a later hosted/platform gate. Version manifests and cross-builds alone cannot check native acceptance.

### U1. Embedded schema and atomic migrations

- [ ] TUSK-V02 **Normal / migration:** On a fresh temporary file, open twice: tasks, task_events, indexes and migration ledger exist once; the second open performs no migration write and preserves inserted data.
- [ ] TUSK-V03 **Failure / migration:** Inject failure after the first DDL/data statement and before ledger commit; then reopen. Old tables/rows/version remain, no partial new schema/event table exists, and a corrected migration can succeed.
- [ ] TUSK-V04 **Compatibility / recovery:** Present a future schema version and an altered applied checksum. Both fail without DDL or journal-mode changes. A down/up cycle on a disposable fixture is reversible; installed-data downgrade is refused.

### U2. Reproducible sqlc queries

- [ ] TUSK-V05 **Normal / generation:** Generate twice from pinned sqlc and identical input. Output is byte-identical; adapter compiles against generated methods. Missing generator, wrong version, or stale output fails the canonical generator gate.
- [ ] TUSK-V06 **Boundary / SQL queries:** Use root, child, 10-level chain, two independent roots and missing root fixtures. Children return immediate children; subtree includes target and all descendants once; nonexistent root produces a not-found error at the repository boundary.
- [ ] TUSK-V07 **Security / query parity:** Use quotes, percent, underscore, non-ASCII case pairs, hyphenated tags, duplicate sort keys and null dates. Bound SQL treats input literally; final results/order match core, including exclusive date bounds and tag intersection.

### U3. Connection and file lifecycle

- [ ] TUSK-V08 **Normal / path:** Run path resolution with override, absolute XDG, relative XDG, unset variables and missing home. Assert exact precedence, literal spaces/#/?/% in filenames, explicit-path-relative-to-cwd behavior and actionable missing-home failure.
- [ ] TUSK-V09 **Security / filesystem:** Create default directory/database with restrictive permissions on POSIX. Test existing parent permissions remain unchanged, wrong file type/symlink at database path, inaccessible directory, and URI-looking input. Reject unsafe targets; never fall back or interpret path text as SQLite options.
- [ ] TUSK-V10 **Concurrency / pragmas:** Force pool connection replacement; read back foreign_keys, synchronous, busy_timeout on each new connection and WAL on disk. Confirm one writer and bounded readers; close/reopen does not lose configuration.
- [ ] TUSK-V11 **Boundary / fixture isolation:** Open two separately named shared-memory fixtures and a disk fixture. Parallel memory tests cannot see each other's rows; final handle close releases memory DB. Memory reports its actual journal mode; only disk WAL may close WAL coverage.

### U4. Repository and transaction contracts

- [ ] TUSK-V12 **Normal / repository:** Round-trip every field, UTC nanoseconds, nulls, empty notes/tags and 255-code-point Unicode title; returned copies do not alias earlier reads or inputs. Canonical IDs persist unchanged.
- [ ] TUSK-V13 **Failure / integrity:** Inject invalid enum/progress/tag JSON/timestamp/UTF-8 rows in a disposable fixture through a controlled test seam. Decode must fail with context; no silent coercion, clamp-and-save, skipped bad row, or fabricated empty list.
- [ ] TUSK-V14 **Failure / transaction:** Within one callback insert a task and event, then inject child/ancestor/event/commit failure, including an ignored statement error. No partial operation is visible. Unknown commit/rollback outcomes require readback rather than an assumed rollback; preserve mapped primary errors and discard unusable connections. Feature 002 V52–V57 owns these cases.
- [ ] TUSK-V15 **Error mapping / recovery:** Exercise missing row, duplicate ID, missing parent, self-parent, children-present, nested transaction, cancellation and busy errors. Assert errors.Is/typed errors, no deadlock, and a subsequent valid transaction succeeds.

### U5. Disk concurrency and recovery proof

- [ ] TUSK-V16 **Concurrency / committed snapshot:** With barriers, hold a read snapshot while another connection commits; the old snapshot remains consistent and a new reader sees the commit. Two fixture callbacks updating distinct children under the same parent preserve both changes and their supplied ancestor results. Phase 3 separately proves the service computes those rollups correctly.
- [ ] TUSK-V17 **Failure / cancellation:** Hold an independent process writer lock. A short-deadline operation returns cancellation before the 5000-ms busy limit within documented scheduling tolerance; an uncancelled contender times out as busy. Release lock and prove the next write succeeds.
- [ ] TUSK-V18 **Recovery / process crash:** Kill a fixture writer at defined points before commit, during a multi-row mutation, and after acknowledged commit. Reopen through Tusk; compare complete old/new graph and event sets, then run quick_check and foreign_key_check. No test unlinks WAL/SHM.
- [ ] TUSK-V19 **Performance / WAL growth:** Hold a reader while writes grow WAL, release it, then verify automatic checkpoint progress and file reopen. Record bounded connection count and query plan. Integrity must hold; this does not claim power-loss durability.

### U6. Application commands and patch contracts

- [ ] TUSK-V20 **Boundary / service input:** Pass missing title, invalid status/priority/tag/progress, invalid UTF-8, absent ID and conflicting patch directives. Reject with expected domain/port errors before beginning a write; source task/draft remains unchanged.
- [ ] TUSK-V21 **Compatibility / patch intent:** Omit due/tags/parent/notes and preserve them; explicitly clear them and observe null/[]/empty values. Empty patch is rejected by CLI; service no-op preserves timestamps and creates no event.

### U7. Deterministic dates and identity

- [ ] TUSK-V22 **Normal / dates:** At fixed local time, parse every accepted token including offset RFC3339. Assert exact UTC instant, weekday-including-today behavior, today end-of-day and tonight 20:00 even after 20:00.
- [ ] TUSK-V23 **Boundary / calendar:** Use leap day, January 31 +1m, DST spring/fall days, malformed offset, invalid calendar date, zero/negative/overflow durations, unknown words and yearless dates. Calendar arithmetic matches R15; errors never fall back to now.
- [ ] TUSK-V24 **Boundary / day query:** Seed tasks exactly at local midnight, one nanosecond before next midnight, exactly next midnight, undated, done and due exactly now. Day query is [start,nextStart); overdue excludes done, undated and due==now.
- [ ] TUSK-V25 **Failure / identity:** Supply deterministic clock/entropy and assert UUIDv7 version/variant/format. Inject entropy error, clock boundary and duplicate DB ID; no task/event persists. Same-millisecond generation does not require global mutable state or ordering assumptions.

### U8. Atomic hierarchy and rollup mutations

- [ ] TUSK-V26 **Normal / rollup lifecycle:** Create root with children 100/0/0 and expect 33. Add/remove/move children, manually update a leaf, complete/reopen descendants, and compare persisted rollup at every level, including 100 from non-done intermediate parents.
- [ ] TUSK-V27 **Boundary / graph validation:** Cover empty forest, one root, depth 2 and 10, rejected depth 11, direct self-loop, two-node loop, deep loop, missing parent, root promotion and subtree-height overflow. Assert sentinel identity; every rejected mutation preserves graph/events.
- [ ] TUSK-V28 **Normal / complete and reopen:** Complete a parent recursively with one clock value; repeat and preserve existing completion times/events. Reopen leaf to 0 and ancestors to in-progress. Reopen parent with all children done and keep parent open at calculated 100.
- [ ] TUSK-V29 **Configuration / auto-completion:** Run identical fixtures with policy false and true, including nested 100-but-open children and explicit parent reopen. Only true with all direct children done auto-completes. An empty child set never triggers auto-completion.
- [ ] TUSK-V30 **Recovery / move and deletion:** Move a subtree between roots with shared ancestors, then delete it. Recompute both old/new chains in dependency order; shared ancestors see final values. Last-child removal resets open parent to 0, done parent to 100, and never restores pre-rollup manual progress.
- [ ] TUSK-V31 **Failure / mutation atomicity:** Inject failure at each changed descendant, ancestor and history append in complete/reopen/move/delete. Readback shows wholly previous state; concurrent opposite moves cannot both succeed and create a cycle.

### U9. Queries, statistics and history orchestration

- [ ] TUSK-V32 **Normal / filtering and ordering:** Compare service results against a fixture oracle for default/explicit status/all, OR enum values, AND tags/fields, root/parent, Unicode search, null dates and ID tie-breaks. List is flat; siblings use independent stable order.
- [ ] TUSK-V33 **Normal / statistics:** Empty DB gives zeros and four status keys. Seed nested parent/child tasks, old/recent completions and overdue tasks; verify R17 formulas and trailing-window boundaries. Reopen/delete updates totals and retained completion metric.
- [ ] TUSK-V34 **Normal / timeline:** Create/edit/status/move/progress/rollup changes append the exact changed-field names and stable sequence; no-op adds nothing. No stored event contains old notes/title. Deletion removes task events and retains unrelated task events.
- [ ] TUSK-V35 **Boundary / tree projection:** Query a selected non-root subtree and filtered TUI context. Display root depth resets to 1 while stored parent stays unchanged; missing ancestors/corrupt cycles error rather than being silently promoted or omitted.

### U10. Service integration and stale-write proof

- [ ] TUSK-V36 **Concurrency / stale form:** Open a form snapshot, commit an external edit to the same editable fields, then submit the form. Return conflict with preserved draft; current committed values/events survive. Changes to unrelated tasks do not falsely conflict.
- [ ] TUSK-V37 **Concurrency / delete confirmation:** Confirm a subtree, insert/move a descendant externally, then execute. Non-forced deletion refuses changed membership; force still requires recursive for parents and acts on the authoritative transaction snapshot.
- [ ] TUSK-V38 **Failure / acknowledged mutation:** Commit a mutation then fail its read-refresh or output step. Report the mutation as committed with refresh/output failure; reread confirms one event set. No automatic replay creates duplicates.

### U11. Lazy CLI routing and process exit

- [ ] TUSK-V39 **Non-regression / lazy startup:** With impossible DB path, invalid runtime config and a temporary home, invoke root/subcommand help, version, completion generation and unknown syntax. All avoid the service factory, storage, network and directory creation; syntax exits 2.
- [ ] TUSK-V40 **Boundary / CLI parsing:** Test no args, unknown commands/flags, missing/excess arguments, negative-looking title after --, malformed numeric flag and validly parsed but invalid domain value. Verify exact 0/1/2 split, stdout/stderr, and usage placement.
- [ ] TUSK-V41 **Configuration / CLI lifecycle:** Test flag-over-environment completion policy, invalid bool config, environment isolation between invocations, cancellation and service close failure. No process-global command state leaks into a second invocation.

### U12. Scriptable mutation commands

- [ ] TUSK-V42 **Normal / command mutation parity:** CLI add/edit/done/reopen via edit performs the same operations as service/TUI; verify due/tag/root clear flags, comma/repeated tags, Unicode titles, empty notes and exact task result.
- [ ] TUSK-V43 **Safety / deletion modes:** Run delete on leaf/parent with recursive/force on/off across TTY, non-TTY and JSON. Noninteractive never prompts; force alone cannot delete children. Default No preserves all rows; approved unchanged subtree deletes exactly intended rows.
- [ ] TUSK-V44 **Boundary / mutation errors:** Empty edits, contradictory flags, invalid status transition, manual progress on parent and nonexistent task fail with correct exit and no mutation; stderr is actionable and does not expose notes.

### U13. List, tree, stats and history commands

- [ ] TUSK-V45 **Normal / query routes:** Run list/tree/stats/history for empty DB, populated DB, existing task without events and missing task. Empty arrays and zero metrics are valid JSON; missing task is an operational error, not an empty success.
- [ ] TUSK-V46 **Compatibility / public semantics:** Exercise explicit status/all ambiguity, root/parent ambiguity, local day filtering, non-root tree projection, history order and full IDs. Confirm query commands do not change task timestamps or history.

### U14. Stable human and JSON formatting

- [ ] TUSK-V47 **Serialization / golden contract:** Decode exactly one JSON value and LF for every data command. Validate field types, UTC times, explicit nulls/[], priority integers and complete status-key map. No ANSI, prompt or human preamble on stdout.
- [ ] TUSK-V48 **Security / human rendering:** Render title/notes containing ESC, OSC hyperlink, CR, tab, newline, wide glyphs and combining characters in TTY/plain mode. No input-controlled terminal command executes; cells fit width and raw JSON round-trips original permitted text.
- [ ] TUSK-V49 **Failure / output writer:** Use a writer failing before first byte and after a partial write. Exit 1, no appended human text on stdout, diagnostics on stderr, and no mutation retry. Distinguish committed mutation from formatting/stream failure.

### U15. Real CLI workflows and latency

- [ ] TUSK-V50 **End-to-end / CLI lifecycle:** In fresh temp home: help → add root/children → list/tree/history → done parent → reopen leaf → move → recursive forced delete → stats. Assert IDs, progress, events, exit streams and persistence after each new subprocess.
- [ ] TUSK-V51 **Performance / startup and queries:** Use the benchmark protocol below for help/version/list/tree/stats/history, including JSON and plain output. Record all raw durations, fixture seed, binary hash/revision, hardware and OS; unmet budgets remain a gate failure.
- [ ] TUSK-V52 **Recovery / real process:** Cancel/terminate a command before and after commit acknowledgment, reopen DB with a new process and inspect state. Repeat failed operation only after authoritative readback; database handles do not prevent later launch.
- [ ] TUSK-V53 **Portability / paths and streams:** Repeat subprocess cases with spaces/non-ASCII path, XDG override, home fallback, LF JSON, closed pipe and non-TTY. Run native OS variants in hosted matrix; local Linux evidence does not establish Windows/macOS success.

### U16. Pure TUI model and deterministic layout

- [ ] TUSK-V54 **Purity / model ownership:** In each loading/empty/loaded/error/modal state call View 100 times, comparing output and deep snapshots of maps/slices/pointers/component state; verify command count remains zero. A hash of serialized exported fields alone is insufficient.
- [ ] TUSK-V55 **Boundary / layout:** Resize 0x0, 1x1, 79x23, 80x24, 120x40, 200x60, back to small/large while editing. No negative dimension/panic/wrap overflow; outer geometry stays fixed across task-count/state changes and draft survives.

### U17. Navigation, filtering and refresh generations

- [ ] TUSK-V56 **Normal / navigation:** Exercise all keys in list/details/empty/filter modes, group collapse, depth-10 scroll, first/last selection and disappearing selected task. Tab focus is visible, selection stays by ID, and all actions remain reachable without color.
- [ ] TUSK-V57 **Race / refresh results:** Deliver older read after newer read, read after successful mutation, and result for closed/replaced modal. Stale messages cannot overwrite current state, reopen modal, or move selection to an unrelated task.
- [ ] TUSK-V58 **Recovery / search and timer:** Use typed fake timers for 150-ms debounce and 2-second refresh; rapidly change/clear search and resize. Cancelled generations are ignored, collapse state restores, failed load preserves labeled stale data, retry succeeds.

### U18. Details, Markdown and timeline

- [ ] TUSK-V59 **Normal / detail viewport:** Scroll long multiline Markdown, tags, history and Unicode at each target size; fields match selected task and CLI history. No selection clears details; terminal output stays within cell bounds.
- [ ] TUSK-V60 **Security / notes:** Feed Markdown with remote image URLs, links, raw controls and malformed formatting. Rendering performs no network/file execution or link launch; safe plain fallback remains readable and storage text is unchanged.

### U19. Forms and confirmed mutations

- [ ] TUSK-V61 **Interaction / modal focus:** Type q, d, ?, spaces and multiline notes in form; they cannot quit/toggle/delete globally. Tab cycles fields, errors stay visible, Esc restores initiating focus, second d cannot confirm delete.
- [ ] TUSK-V62 **Failure / save lifecycle:** Submit twice rapidly; exactly one write runs. Inject validation/storage/conflict failure and retain draft. Commit then fail refresh: display saved/stale state and never resubmit. Small-terminal resize preserves draft during save.
- [ ] TUSK-V63 **Concurrency / TUI mutations:** External CLI edit/delete/move while TUI form/confirmation is open produces conflict or safe missing-task state. Reload/reapply is explicit; latest committed graph/events remain intact.

### U20. TUI workflow and real terminal acceptance

- [ ] TUSK-V64 **End-to-end / synthetic TUI:** Run loading → create → expand → edit due/progress → complete → reopen → external conflict → recover → delete → quit through synthetic messages. Assert service requests, state, focus and output at each boundary.
- [ ] TUSK-V65 **Performance / TUI:** Benchmark bounded loaded/model and Markdown cases under make bench-tui; record allocations and p50/p95/max render times at 1000 tasks and target dimensions. View remains pure; target p95 <16ms is measured, not assumed zero allocation.
- [ ] TUSK-V66 **Manual / terminal lifecycle:** On Linux/macOS/Windows terminals verify keyboard input, resize, Unicode/plain rendering, visible focus and error recovery; cancel/quit/panic-handled startup failure restores echo/cursor/alternate screen. Launch again and verify normal shell behavior.

### U21. Hosted platform quality gates

- [ ] TUSK-V67 **Hosted / canonical gates:** On the candidate SHA, run setup, validate, generated-diff check and native build on Linux/macOS/Windows with explicit Bash/Make/compiler prerequisites. No CGO=0 race job; formatting-induced source diff is a failure.
- [ ] TUSK-V68 **Hosted / target evidence:** Run produced binaries on each supported OS/architecture or record unexecuted architecture as pending. A cross-compiled artifact existing is not runtime acceptance. Attach hosted URLs and exact hashes/revision.

### U22. Release artifacts and installation lifecycle

- [ ] TUSK-V69 **Packaging / artifact manifest:** Snapshot includes each required target, checksum, version and license/notice metadata. Binary runs with CGO-free runtime prerequisites. Invalid config/missing architecture fails release-check; snapshot performs no publication.
- [ ] TUSK-V70 **Recovery / upgrade and downgrade:** Create real fixture data with current binary, replace binary and migrate, then reopen and compare tasks/events. Older binary refuses a newer schema without changing files. Removal of executable leaves data intact; restore backup only using documented consistent procedure.
- [ ] TUSK-V71 **Manual / release readiness:** Verify destination, ownership, notices, Homebrew formula/source checksum if used, documented install/upgrade/remove and release candidate evidence. Keep publication pending until an authorized release action actually succeeds.

### U23. Completions, documentation and final handoff

- [ ] TUSK-V72 **Compatibility / completions:** Generate Bash/Zsh/Fish completions and man pages with unusable DB config and no data directory. Parse/load in each real shell, exercise representative command/flag completions, and prove generation requires no storage.
- [ ] TUSK-V73 **Documentation / handoff:** Replay documented CLI/JSON/delete/date/backup examples against candidate artifacts. Cross-link product and feature triplets, require all scenario/issue evidence and exact revision in master checklist, and leave pending hosted/manual/publication claims explicit.

## Commands and environments

| Tier | Canonical command or method | Availability and planned evidence |
| --- | --- | --- |
| Environment | `make setup` | Exists; current toolchain check, not proof dependency minimums are compatible |
| Focused | `make test-unit` / `make test` | Exists; assert the named tests ran; no supported TEST/PKG/RUN variables at baseline |
| Aggregate | `make validate` | Exists; fmt/vet/test/race/coverage; record package coverage and exclusions |
| Concurrency | `make race` | Exists; ensure disk/process tests are included, not skipped as short |
| Binary | `make build` | Exists; native build; subprocess fixtures invoke this once through Make |
| Core regression | `make bench-tree`, `make bench-build` | Exists; historical core benchmarks, not CLI latency proof |
| SQL generation | `make generate`, `make check-generated` | Planned in U2; not executable now |
| Storage compatibility | `make test-compat`, `make build-storage` | Planned in product U24 / Feature 002 U6; minimum compiler and cross-build evidence |
| SQL tool and scripts | `make setup-sqlc`, `make test-scripts` | Planned in U2; explicit pinned tool installation and negative fixtures |
| Storage cost | `make bench-storage` | Planned in U5; does not prove CLI latency |
| CLI process/performance | `make test-cli`, `make bench-cli` | Planned in U15; not executable now |
| TUI performance | `make bench-tui` | Planned in U20; not executable now |
| Release | `make release-check`, `make release-snapshot` | Planned in U22; snapshot does not publish |
| Shells | `make test-completions` | Planned in U23; native Bash/Zsh/Fish parse/load evidence |
| Hosted | Candidate-SHA matrix using canonical Make targets | U21; store job URLs, OS/arch, compiler and artifact hash |
| Manual | Real-terminal scripts in V66/V71 | Store terminal/OS/version, dimensions, actions, observations and artifacts |

## Fixtures and failure injection

- Use a unique temporary home/data directory and file per test. No developer database is touched. Memory URIs include a per-test unique name; close every reader/writer before cleanup.
- Freeze clock and location. Include UTC, America/New_York DST transitions, leap years/month ends, dates on each interval boundary, missing due dates, Unicode/code-point/cell-width cases, nil/empty fields, 10-level trees, corrupt/orphan/cyclic fixtures, and duplicate IDs.
- Fakes must model transaction rollback and fail at begin/read/write/event/commit/rollback/close boundaries. Real SQLite tests prove behavior fakes cannot: lock acquisition, committed snapshots, foreign keys, generated queries, persistence and crash recovery.
- Use barriers/channels/test-process handshakes to position concurrency; do not use arbitrary sleeps to assert ordering. Every test has a bounded deadline and cleans up child processes.
- Fixture-only crash control identifies pre-commit and post-acknowledgment points. Never install a crash hook in normal runtime behavior. A killed process test does not simulate sudden hardware power loss.
- Permission tests must detect privileged runners that bypass permission bits and use a deterministic filesystem-error seam; a false success as root is not coverage.
- Use sentinel note/title strings to detect accidental history/log retention. Corruption bypasses and invalid SQL are confined to fixture databases.
- TUI deep snapshots must include child component state, slices/maps/pointers, draft, focus, collapse state and request generation. Check read purity with repeated output and state comparisons, not only a hash of exported fields.

## Performance protocol

U15's canonical benchmark runner measures a freshly launched compiled binary per sample with stdout drained to a sink; build time and shell startup are excluded. Capture hardware/CPU, OS, filesystem, toolchain, exact binary hash/revision and all raw durations.

Fixtures: empty initialized DB, 100 tasks, 1000 tasks (128-character notes, three tags, a ten-level branch), and a 10000-task stress fixture. Run help/version with invalid DB configuration as well as a clean environment. Query list/tree/stats/history in plain and JSON modes; choose a root/history fixture with a documented output size. Record first-ever database creation/migration separately.

For each command/fixture use 100 fresh processes after five warmup invocations; preserve warmup results separately. Report p50/p95/max and number exceeding the mandate. On the declared reference environment, every measured normal sample must be below 5 ms for help/version and 15 ms for queries. Do not discard outliers or silently redefine a failed maximum as a p95 success. G3 selects and records the reference environment before measurement. The 10000-task stress fixture and slow-pipe/lock-contention runs expose scaling and exception behavior; do not claim those pass a latency gate unless measured.

The absolute project wording cannot establish a universal bound for arbitrary hardware, output volume or contention. Record any observed violation with fixture/environment in the workorder; optimize or obtain an explicit revised product acceptance contract, never mark it green through a narrower benchmark alone.

U20 reports View/render times at 80x24, 120x40 and 200x60 on the loaded fixture, with allocations and p50/p95/max. The research target is p95 below 16 ms; View's correctness gate remains purity even if allocating output strings is necessary. No zero-allocation claim is inherited from old research.

## Execution record

All TUSK-V01–TUSK-V73 remain **not executed**. Planning does not populate software pass/fail rows. Existing Phase 001 results remain in that feature's original triplet and were not rerun here.

| Date | Exact revision and dirty scope | Scenario/test | Environment | Command/method | Result | Evidence location |
| --- | --- | --- | --- | --- | --- | --- |
| — | — | — | — | — | Not executed | — |

For a red/green unit record both failures and results separately. Attach local command logs and hosted/manual proof by exact revision; if sources change after a result, identify which evidence remains applicable and rerun the affected checks. Record documentation-audit results in the workorder's planning record, not this software execution table.

## Feature 002 local handoff — 2026-09-08

The [Feature 002 triplet](../plans/2026-09-06-002-feat-sqlite-storage-and-repository-plan.md) now records six implemented units with per-unit validated commits and 67/68 local scenarios accepted. [Storage operations and Phase 3 obligations](../storage.md) and [durable evidence](../verification-evidence/002/README.md) cover the repository boundary, atomic metadata history, migration refusal, process recovery and benchmarks. Product U24 compatibility evidence is available locally; target-native TUSK-V01 acceptance remains pending. This handoff does not check cross-phase service/CLI/TUI or hosted/native product scenarios. At that handoff MASTERPLAN.md advanced to Feature 003 planning. Its completed planning pack is recorded in the subsequent Feature 003 handoff below.

## Feature 003 planning handoff — 2026-09-09

The [Feature 003 plan](../plans/2026-09-06-003-feat-task-service-engine-plan.md), [verification matrix](../verification-plans/2026-09-06-003-feat-task-service-engine-verification-plan.md), and [workorder](../workorders/2026-09-06-003-feat-task-service-engine-issues-workorder.md) now satisfy G1 for service planning. Product U6 → feature U1, U7 → U2, U8 → U3/U6/U7, U9 → U4, and U10 → U5. The service pack has 22 requirements, seven units and 91 unexecuted scenarios; all 73 product scenario IDs/check states remain unchanged.

TUSK-V20–V38 service coverage is expanded in that matrix. TUSK-V35 filtered TUI presentation and TUSK-V38 post-commit output/refresh presentation still need the owning adapters. The new service workorder separately tracks failed-rollback error-category compatibility (U1), real disk atomicity/recovery and runtime/benchmark proof (U5), and production timezone-data/CLI/TUI/native/hosted handoffs (Features 004–006). No product requirement or runtime/release checkbox closes in this planning update. The Feature 003 workorder owns the current documentation audit; no application tests or make validate ran.


## Feature 004 planning handoff — 2026-09-09

The [Feature 004 plan](../plans/2026-09-06-004-feat-cli-interface-and-scripting-plan.md), [verification matrix](../verification-plans/2026-09-06-004-feat-cli-interface-and-scripting-verification-plan.md) and [workorder](../workorders/2026-09-06-004-feat-cli-interface-and-scripting-issues-workorder.md) now satisfy G1 for CLI planning. Features 002/003 are locally accepted and merged according to MASTERPLAN.md; earlier planning handoffs above are historical.

Product U11→feature U1, U14→U5/U4, U12→U2/U7, U13→U3 and U15→U6. Execution order is product U11→U14→U12→U13→U15 / feature U1→U5→U4→U2→U7→U3→U6. Existing product IDs and all 73 scenario check states are preserved. Feature 004 has 25 requirements, seven units, 91 unchecked scenarios and 25 findings; 20 are fixed in planning and five remain execution/release evidence gates.

TUSK-V39–V53 expand into that matrix, with TUSK-V38 output recovery and TUSK-V73 governance included. G3 CLI dependency choices are recorded in Feature 004 KTD1; U1 still must prove the combined graph/Go 1.25/full executable builds, and U6 must meet the unchanged product performance protocol. V90–V91 retain native/hosted release obligations with Feature 006. G1 stays open for Features 005/006. No application tests, dependency builds or make validate ran, and no runtime/release checkbox closes during planning. MASTERPLAN names Feature 004 U1 as next, awaiting implementation authorization.
