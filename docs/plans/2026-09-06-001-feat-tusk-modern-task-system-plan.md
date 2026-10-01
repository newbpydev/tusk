---
artifact_contract: ce-unified-plan/v1
artifact_readiness: requirements-only
product_contract_source: ce-brainstorm
feature-id: TUSK
title: Tusk Modern Task Management System - Plan
type: feat
date: 2026-09-06
deepened: 2026-09-08
planning_scope: product-contract-and-cross-phase-handoffs
---

# Tusk Modern Task Management System - Plan

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


## Goal Capsule

Give developers a local task manager usable immediately from a terminal, shell script, or AI agent, with a keyboard-driven TUI over the same task operations. Use one binary, embedded SQLite, a pure Go domain core, and synchronous services called by CLI handlers or Bubble Tea commands.

- **Authority:** [AGENTS.md](../../AGENTS.md) governs engineering; [MASTERPLAN.md](../../MASTERPLAN.md) alone governs execution status and order.
- **Scope:** Product contracts and ordered handoffs for Features 002–006; Feature 001 is the existing compatibility baseline.
- **Surfaces:** CLI/TUI, internal library/service, persistence/migration, packaging/operations, documentation.
- **Artifact triplet:** This plan, its [verification plan](../verification-plans/2026-09-06-001-feat-tusk-modern-task-system-verification-plan.md), and [issue workorder](../workorders/2026-09-06-001-feat-tusk-modern-task-system-issues-workorder.md), under `docs/` in this first-party repository.
- **Readiness:** Deepened product planning baseline with explicit gates. This umbrella remains `requirements-only` as an execution entrypoint. It does not replace the phase-specific triplets or authorize implementation.
- **Current handoff:** Features 002–005 are locally accepted; local/remote main records PR #5 merged at `42d7c52` and cleanup at `2f2d3ff`. Product U16–U20 retain their local acceptance. [Feature 006](2026-09-06-006-feat-automation-packaging-and-release-plan.md) now has a synchronized planning pack with 30 feature requirements, eight units and 102 unexecuted scenarios. Product G1 planning is satisfied for Features 002–006; native/hosted/publication G4, including remaining platform coverage of TUSK-V66, remains open. No Feature 006 implementation or hosted action is authorized by this planning checkpoint.

## Product Contract

### Summary

Tusk provides recursive task hierarchies, progress rollup, tags, due dates, CLI automation, Markdown notes, and an interactive terminal workspace. Local use requires no daemon, network account, or database service.

### Problem Frame

The legacy application required PostgreSQL during startup, prompted for authentication in batch flows, and mutated TUI state during rendering. These behaviors prevented reliable scripting and caused stale or unstable terminal views. The [legacy audit](../research/legacy-audit.md) supplies regression cases; its implementation is archived, not a template for reuse.

### Preservation and live baseline

Product Contract restructured: original sections 2–5 map to R1–R27 below. Local operation, CLI/TUI identity, task fields, Markdown notes, optional parent completion, timeline history, and release platforms are retained. Clarifications reconcile original prose with shipped core behavior: depth 10, manual leaf progress, hyphenated tags, and exact state transitions. “Collision-free” becomes collision detection; “sub-millisecond” prose does not override the latency mandates.

Historical product-planning baseline (superseded by the current handoff): `6128d312921cecca2a024bc8647d959126822f0b` on `main`, 2026-09-08; worktree initially clean. `go.mod` declares Go 1.24.0 and no dependencies. Only `internal/core/` and a help/version scaffold in `cmd/tusk/` exist. Ports, storage, services, Cobra, TUI, schema, and hosted workflows do not exist. Phase 0/1 completion is recorded historical evidence in `MASTERPLAN.md`, not a fresh test result from this pass.

### Actors and flows

- A1. Developer: capture tasks, manage subtasks, review due work, and edit notes locally.
- A2. Script/AI agent: perform the same durable actions through explicit arguments, JSON, stable IDs, and no hidden prompt.
- A3. Maintainer: migrate data, reproduce defects, validate releases, and distribute binaries.
- F1. Capture: parse → validate metadata/date → create task and update ancestors atomically → return committed ID.
- F2. Mutate: read authoritative graph under a write transaction → validate → change affected descendants/ancestors and history → commit → refresh.
- F3. TUI: load through a command → navigate snapshot → submit form → accept matching result → reload authoritative data.
- F4. Upgrade: lazy open → check compatibility → migrate atomically if needed → serve command; failure preserves prior data.

### Requirements

#### Local runtime and task model

- R1. Local commands require no network, authentication, service, or daemon. Help, version, completion generation, and syntax errors never initialize storage or create directories.
- R2. Storage precedence: nonempty `TUSK_DB_PATH`, absolute `XDG_DATA_HOME/tusk/tusk.db`, then home plus `.local/share/tusk/tusk.db`. Relative explicit paths resolve against invocation directory; relative XDG values are ignored. The override is a filesystem path, never a user-controlled DSN. Invalid/unwritable paths return an operational error without silently choosing a different database.
- R3. Retain ID, title, description, status, priority, parent ID, progress, tags, due date, creation/update/completion timestamps. Core IDs remain opaque nonempty strings; service-created IDs use canonical lowercase UUIDv7. No prefix matching or automatic duplicate-insert retry.
- R4. Title: trimmed, 1–255 Unicode code points. Defaults: `todo`, `medium`, progress 0, empty notes, null optional fields. Tags use `core.NormalizeTags`: optional `#`, lowercase ASCII, single internal hyphens, 1–32 characters, sorted/deduplicated. Public adapters reject invalid UTF-8 and NUL without changing valid stored text.
- R5. Preserve the core state machine: open states can transition to any valid state; `done` reopens to `todo`/`in-progress`, not directly to `blocked`. Completion stamps UTC time; reopening clears it. Repeated satisfied commands do not change timestamps or create duplicate events.
- R6. Trees are acyclic, root depth 1, maximum 10 levels. Moves include descendant height in the depth check. Missing parents fail. Self-parenting matches both `ErrSelfParenting` and `ErrCyclicDependency` through `errors.Is`. No implicit orphan promotion.
- R7. Open leaf progress is manual 0–99; done is 100. Open parent progress is the floor average of immediate children; a non-done child parent with rollup 100 contributes 100. Reject manual progress on a parent. Losing the last child resets an open parent's progress to 0 and retains 100 for done; no restoration of historical manual progress.
- R8. Explicit completion marks target and every descendant done atomically. Reopening a leaf resets manual progress to 0. Reopening a parent preserves descendants and recalculates progress, possibly leaving it open at 100. Introducing/reopening an incomplete child below a done ancestor reopens that ancestor to `in-progress` before rollup.
- R9. Automatic parent completion defaults to false. If enabled, a mutation leaving a nonempty set of direct children all `done` completes that parent upward; rollup 100 alone does not trigger it. Explicit reopen wins for the targeted task during that operation. `--auto-complete-parent` overrides `TUSK_AUTO_COMPLETE_PARENT`; TUI receives the same policy. Invalid configuration returns 1; help remains usable.

#### Persistence, queries, and dates

- R10. A mutation's graph validation, task changes, ancestor rollups, and history appends share one transaction. Readers see a committed snapshot. No success before commit and no partial success on ancestor/event failure.
- R11. Embedded sequential migrations run on first data access. Reject unknown newer schema versions and changed applied checksums. Failed migration preserves the previous schema/data; never delete or recreate a database to recover from an error.
- R12. Disk databases use WAL, foreign keys ON, busy timeout 5000 ms, synchronous NORMAL. Lock waits and cancellation return errors when bounds expire. Disk fixtures prove WAL/concurrency/recovery; memory fixtures prove repository semantics only.
- R13. Filters OR within statuses/priorities and AND across fields; all requested tags must match. Search matches `core.FilterTasks` literal case-insensitive title/description substring semantics. Default list excludes done; explicit status selects exactly those states; `--all` includes done. Combining explicit status and `--all` returns 2.
- R14. Default order: priority descending, due ascending with null last, creation ascending, ID ascending. Apply to siblings independently. The glossary's pinned/active ordering is reserved terminology, not a persisted pin capability in this release.
- R15. Parse against one injected reference time and local zone; persist UTC. Support `today`, `tomorrow`, `tonight`, weekday abbreviations, positive `+Nd`, `+Nw`, `+Nm`, `YYYY-MM-DD`, and RFC3339 with offset. Day tokens/bare dates end at 23:59:59.999999999 local; `tonight` means 20:00 today even when past. Weekdays mean next occurrence including today. Offsets use calendar days/weeks; months clamp to last valid day. Reject overflow, invalid dates, and yearless `Sep 15`; no parse-error fallback to now.
- R16. `list --due <date>` selects the parsed local day with inclusive start and exclusive next-day boundary; undated tasks are excluded. Overdue means non-done and due strictly before reference time. Preserve core's exclusive `DueBefore`/`DueAfter`; the service owns the separate half-open day interval.
- R17. Statistics count all retained tasks, including parents: total, counts by status, done count, floor(done/total × 100), overdue count. Empty completion percent is 0. “Completed last 7 days” counts currently done tasks with `now - 168h < completed_at <= now`. Reopening/deleting removes them from that metric; do not describe it as immutable historical productivity.
- R18. Details timeline records create, metadata edit, status, move, manual-progress, and rollup changes with UTC time and stable sequence. Store event kind and changed field names, not old notes/titles or snapshots. Deleting a task deletes its timeline; no tombstones/undo. Expose it through CLI `history <id> --json` and the TUI.

#### CLI and machine contract

- R19. The grammar below covers every durable TUI action without a TTY. Omitted edit fields preserve values; explicit clear flags clear fields. Empty edits and contradictory clear/set flags return 2. A compound edit cannot combine manual progress with status/parent changes.
- R20. Parent deletion requires `--recursive`. Without `--force`, a TTY gets default-No confirmation naming target and subtree count; non-TTY/JSON fail with guidance. Force skips confirmation only, never implies recursion. Decline succeeds without mutation. Recheck the confirmed subtree under the writer lock; changed membership requires new confirmation unless forced.
- R21. `add`, `list`, `done`, `edit`, `delete`, `tree`, `stats`, and `history` support `--json`. Emit one compact UTF-8 JSON value plus LF on stdout; diagnostics on stderr. Pre-output errors leave stdout empty. Output failure never appends human text to JSON or replays a mutation.
- R22. Exit 0 success, 1 operational/domain/configuration/output failure, 2 syntax/unknown command/flag/arity. Cancellation before commit returns 1 without success. If commit succeeded before cancellation/output failure, retain committed state and require readback before retrying a mutation.
- R23. Human output omits ANSI when non-TTY, `NO_COLOR` is nonempty, or `TERM=dumb`. Stored text cannot inject terminal controls or OSC hyperlinks. JSON preserves text through JSON escaping. TUI requires terminal stdin/stdout; empty invocation shows help and does not start TUI.

#### TUI and delivery

- R24. Use the Bubble Tea v1 contract `View() string`. Constructors initialize components; `Update` changes state/layout/styles. Async loads, timers and writes are `tea.Cmd` with typed messages. View never mutates any model, pointer, slice, map, style, component, or cache.
- R25. States: loading, empty, filtered-empty, loaded, refreshing, load error, saving, save error, stale data. Preserve selection by ID; reject superseded responses. Failed writes preserve drafts; uncertain commits require reload before retry. A late response cannot overwrite newer data or a different form.
- R26. At least 80×24 supports list/details/help without overflow. Smaller terminals show a bounded resize message and allow quit, preserving drafts/selection. Support visible focus, non-color status labels, Unicode cell widths, scrolling, keyboard operation, and plain terminal presentation. Real terminal acceptance is separate from synthetic tests.
- R27. Release CGO-free binaries for Linux amd64/arm64, macOS amd64/arm64, Windows amd64, Bash/Zsh/Fish completion, and man pages. Keep local, hosted, manual, and publication evidence separate. Data survives executable replacement/removal.
- R28. Preserve reference query p90 below 15 ms and help/version p90 below 5 ms, with the current tail guards. Measure actual subprocess startup through exit, including formatting, using explicit fixtures. Report initialization, lock contention, slow output, and larger datasets separately; no benchmark proves the bound on all hardware or unbounded output.
- R29. Core stays independent; state is constructor-injected; verification uses Makefile gates. `make validate` currently includes fmt, vet, test, race, and at least 95% per-package coverage for nonexempt packages. Close a phase only after its scenarios, issue gates, and master checklist have execution evidence.
- R30. GitHub repository metadata and README give users accurate installation/use instructions, verified platform/version support, evidence-backed screenshots/performance claims, and working contributor/security/license links. Publish only available installation routes and inspect actual rendered metadata/README after authorized changes. Owner-requested addition on 2026-10-01; Feature 006 owns execution.

### Command grammar and data shapes

| Command | Arguments and flags | Result |
| --- | --- | --- |
| `add <title>` | `-p/--priority`, `-d/--due`, repeatable `-t/--tags`, `--parent`, `-n/--notes`, `--json` | Created Task |
| `list` | repeatable `-s/--status`, `-p/--priority`, `-t/--tags`; `--due`, `--search`, `--parent`, `--root`, `--all`, `--json` | Flat array |
| `done <id>` | `--json` | Selected Task after subtree completion |
| `edit <id>` | `--title`, `--notes`, `--priority`, `--status`, `--progress`, `--due`/`--clear-due`, repeatable `--tags`/`--clear-tags`, `--parent`/`--root`, `--json` | Task after atomic patch |
| `delete <id>` | `--recursive`, `--force`, `--json` | Delete result |
| `tree [id]` | `--json` | Forest or subtree, including done descendants |
| `stats` | `--json` | R17 metrics |
| `history <id>` | `--json` | Oldest-first event array |
| `tui` | common policy flag | Terminal session |
| `help`, `version`, `completion <shell>` | root `--help`/`--version` aliases | No storage |

Unknown flags, extra arguments, simultaneous `--root`/`--parent`, malformed numeric/boolean syntax return 2. Invalid domain values (status, date expression, out-of-range parsed progress) return 1. Titles beginning with `-` can follow `--`. Tags flatten repeated/comma-separated values; empty elements fail. Explicit empty notes clear notes. `edit --status done` uses subtree completion; `--status todo/in-progress` on a done task uses reopening. In a parent+status patch, validate/move first, apply status, then recalculate both ancestor chains atomically.

Wire DTOs are independent of core's existing `omitempty` tags:

- Task: `id`, `title`, `description` (including empty), `status` string, `priority` integer 1–4, `parent_id` string/null, `progress` integer, `tags` array (empty `[]`), `due_date`, `created_at`, `updated_at`, `completed_at` UTC RFC3339Nano strings; optional dates null.
- List: array, empty `[]`. Add/done/edit: one Task. Tree: array of `{task, children, depth}` with children `[]`; selected subtree root has display depth 1 without changing stored parent ID.
- Delete: `{id, deleted_ids, deleted_count, deleted}`; sorted IDs. JSON requires force and never prompts.
- Stats: `{total, by_status, done, completion_percent, overdue, completed_last_7_days}`, with all four status keys.
- History: array of `{sequence, task_id, kind, changed_fields, occurred_at}`; nonexistent task fails, existing task with no events returns `[]`.

No pagination or implicit JSON truncation in this release. Additive fields are compatible; field removal/type/null/array changes require an explicit compatibility revision and fixtures.

### Scope and assumptions

No network sync, external backend implementation, accounts, recurrence, reminders, collaboration, non-parent task dependencies, or event-sourced reconstruction. A future backend may implement the same port; no plugin framework is needed.

Planning defaults are R9 opt-in completion, R15 date times, R17 retained-task metrics, R18 metadata-only timeline, and R20 explicit recursive deletion. They are decisions for this pack, not claims of separate user approval. Changes require synchronized product and affected feature triplets. Timeline/history is retained original scope; Feature 002 now supplies its persistence, and Feature 003 owns event selection and service exposure.

## Planning Contract

### Technical decisions

- KTD1. **Phase plans own execution.** Units below are handoffs. Features 002–006 have complete planning packs; Features 002–005 have local execution evidence. Feature 006 has no execution evidence yet and owns remaining native/hosted/publication proof. Gate G1 applies per feature. Follow the masterplan sequence even where the DAG allows concurrency.
- KTD2. **Retain the accepted modernc SQLite runtime.** Keep `database/sql` in storage. Feature 002 records local proof for the matching engine/driver/libc/Go 1.25 baseline; native/hosted G2/G4 acceptance remains pending. The Feature 003 planning pass changes no dependencies.
- KTD3. **Migration-owning `db` package.** `db/embed.go` embeds adjacent `migrations/*.sql`; storage imports it. Embedding `../../db` from storage is invalid. Generate queries into `internal/storage/sqlc/`; test generated behavior through the adapter. [Go embed](https://pkg.go.dev/embed), [sqlc configuration](https://docs.sqlc.dev/en/latest/reference/config.html).
- KTD4. **One writer, bounded readers.** Use one writer connection and up to four reader connections. Begin writes IMMEDIATE before reading graph state; read snapshots use separate deferred transactions. Configure connection-scoped pragmas for every new/replacement connection. No transaction callback may borrow another write connection. Verify cancellation and driver transaction options at G2. Create new application-owned directories as 0700 and database files as 0600 on POSIX; preserve existing parent permissions and use the user's private profile ACL on Windows. Reject symlink/non-regular database targets; encode path characters as literal filename data when constructing an internal DSN. This is a single-user boundary, not protection against a malicious process with the same OS identity. [Driver documentation](https://pkg.go.dev/modernc.org/sqlite).
- KTD5. **Transaction-scoped ports.** `ports.TaskRepository` exposes reads and a write callback receiving a restricted query/writer interface. The callback uses that interface for all reads/writes/events and cannot retain or nest it. Storage owns begin/commit/rollback/close; service owns business decisions. Wrap failures with context and map known domain sentinels via `errors.Is`; typed port errors cover busy, conflict, children-present, corruption, and incompatible schema.
- KTD6. **Strict schema and decoded validation.** Tasks: non-null primary key, enum/range checks, nullable parent FK, self-parent check, done/progress/completion consistency, valid JSON tag arrays. Store UTC timestamps as fixed-width TEXT with nine fractional digits for chronological comparison. Cross-row cycle/depth/leaf/rollup rules remain service checks inside the write transaction. Corrupt decoded rows fail; never clamp and save them.
- KTD7. **Minimal history table.** `task_events`: integer sequence, task FK with cascade deletion, kind, changed-field JSON array, occurred-at. Index task/sequence. No old user text or state snapshots. Tasks remain authoritative. Parent FK also cascades, but service authorizes recursive deletion under R20.
- KTD8. **Query parity.** Index parent, status, priority, and due date; add composites only from query-plan evidence. Bind values; allowlist sorting. SQL narrows candidates; use core filtering for Unicode search rather than assuming SQLite NOCASE equals Go case conversion. Sort through core. Never pass a filtered list as a complete forest to `BuildTree`; fetch complete ancestry/subtree before projection.
- KTD9. **Ledger and recovery.** Acquire writer lock, re-read migration ledger, apply all pending migrations and checksums in one transaction. Current schemas need a compatibility read, not a migration write on each query. Check newer schemas before persistent journal changes. Canonical `db/migrations/NNN_name.sql` files contain forward SQL only; inverse fixture scripts live under `internal/storage/testdata/migrations/` and are never embedded or fed to sqlc. Down migrations are only for disposable test fixtures; installed data uses roll-forward or consistent offline-backup restoration. Keep the original and WAL sidecars on failure.
- KTD10. **Bound durability claims.** NORMAL is retained, without promising the last acknowledged transaction survives power loss. App termination, transaction rollback, reopen and integrity each have scenarios. SQLite owns WAL cleanup; never unlink WAL/SHM to repair data. Require suitable local filesystem semantics. [SQLite WAL](https://www.sqlite.org/wal.html), [pragmas and in-memory limits](https://www.sqlite.org/pragma.html).
- KTD11. **Inject identity/time.** One clock value per mutation, explicit location for dates. A small service helper constructs RFC 9562 UUIDv7 from supplied time and cryptographic randomness, without package mutable state. Entropy/duplicate failures abort; no retry after uncertain commit. Sort by R14, not assumed monotonic UUID order. [UUIDv7](https://www.rfc-editor.org/rfc/rfc9562.html#section-5.7).
- KTD12. **Prevent stale overwrite.** Apply only supplied patch fields to the latest task under writer lock. TUI forms compare their base editable fields to authoritative values; differences return conflict and retain draft. Confirmation compares subtree membership and target metadata. CLI patches operate on the latest row; no long-lived stale row replacement.
- KTD13. **Preserve v1 UI family.** Feature 005 KTD1 selects Bubble Tea v1.3.10, Lipgloss v1.1.0, Bubbles v0.21.0, Glamour v0.9.1 and x/ansi v0.10.1 on the existing Go 1.25 baseline. U1 must prove the combined graph; no silent v2 View/key switch. [Selected pack and versioned sources](2026-09-06-005-feat-interactive-tui-application-plan.md#dependency-evidence).
- KTD14. **Refresh through commands.** Load at start, after mutation, on `r`, and every 2 seconds through a typed timer command. Use request generations, invalidate pending reads after writes, allow one write at a time, preserve drafts. Failed refresh keeps the snapshot labeled stale. No service cache/goroutine pool.
- KTD15. **Separate release from race builds.** CGO=0 for binaries; race gates retain compiler support. Document Bash/Make on Windows and check for formatting-induced diffs in CI. Keep immutable release version metadata using generated constants, not a mutable package variable.

### Architecture and transaction flow

Port operation contract (all I/O accepts context; names below are the planned API vocabulary, with final Go types owned by the phase-specific pack):

| Boundary | Operations | Ownership and result |
| --- | --- | --- |
| Repository reads | GetByID, List, ListChildren, GetSubtree, GetAncestors, ListEvents | Detached core tasks/event read models; missing singular task is ErrTaskNotFound; empty collections are non-nil |
| Repository read snapshot | WithRead | One callback sees a consistent task/event snapshot across multiple reads; no mutation methods exposed |
| Repository write transaction | WithWrite | Callback reads plus Create/Update/Delete/AppendEvent; callback/commit failure returns no success; nested callback rejected |
| Service mutations | CreateTask, UpdateTask, CompleteTask, ReopenTask, DeleteTask | Structured commands with supplied/clear intent, optional base snapshot/confirmation, one operation timestamp, committed result |
| Service queries | ListTasks, GetTaskTree, GetStats, GetTaskHistory | Validated filters, snapshot-consistent DTO/read model; tree root projection never changes stored task |

Do not expose `sql.DB`, `sql.Tx`, driver errors or generated query structs through ports. Delete preview and execution carry the exact sorted task IDs and target editable metadata for R20's equality check. Constructors own resource dependencies; only the composition root closes the opened storage instance. Feature 003 pins standard-library time/tzdata for named-zone test fixtures; Feature 004 embeds it in the production main package before date commands ship. G3 retains that packaging proof and CLI/UI dependency decisions.

These sketches define ownership and sequencing, not implementation code.

```mermaid
flowchart TD
  Boot[cmd/tusk: parse and wire] --> CLI[internal/cli]
  Boot --> TUI[internal/tui]
  CLI --> Service[ports.TaskService and internal/service]
  TUI --> Cmd[tea.Cmd and typed result]
  Cmd --> Service
  Service --> Core[internal/core: pure invariants]
  Service --> Repo[ports.TaskRepository]
  Repo --> Storage[internal/storage and generated sqlc]
  Storage --> Embedded[db: embedded migrations]
  Storage --> SQLite[local SQLite file]
```

```mermaid
sequenceDiagram
  participant P as CLI or TUI command
  participant S as Task service
  participant T as Transaction-scoped repository
  P->>S: Intent and context
  S->>T: Acquire writer; BEGIN IMMEDIATE
  S->>T: Read task, subtree and ancestors
  S->>S: Validate conflict, cycles, depth and lifecycle
  S->>T: Write changed tasks and events
  alt failure before commit
    T-->>S: Rollback and wrapped error
    S-->>P: No confirmed mutation; retain input
  else commit succeeds
    T-->>S: Commit acknowledgment
    S-->>P: Committed result
    P->>S: Refresh snapshot
  end
```

### Mutation decisions

| Trigger | Target/subtree | Ancestors and events |
| --- | --- | --- |
| Add child | Validate parent/depth; create default-open child | Reopen done ancestors if incomplete; roll up deepest first |
| Complete | Target/descendants done with one operation time; existing done rows no-op | Recalculate ancestors; auto-complete only per R9 |
| Reopen leaf | Requested open status, progress 0, completion cleared | Reopen done ancestors; preserve siblings |
| Reopen parent | Preserve descendants, clear completion, recalculate progress | Suppress auto-completion of explicit target in this operation |
| Move/root promotion | Validate subtree/cycle/final depth; move once | Recalculate old/new chains deepest first; shared ancestors once after both branches |
| Delete | Recheck confirmation; delete authorized tasks/events | Recalculate old ancestors; apply last-child R7 |
| Metadata/progress | Atomic patch; reject parent manual progress | Roll up when relevant; append actual changed fields only |

### TUI interaction contract

Reserve borders and a bottom help/status row. Use about 40% remaining width for list and 60% for details. Compute nonnegative inner cell dimensions on resize. Empty/error/loading/task-count changes keep outer geometry fixed. Clip by terminal cells; scroll ten-level trees, long notes, and wide glyphs.

Root groups: Today (includes overdue open roots), Upcoming (future due roots), Backlog (undated open roots), Completed (done roots). Descendants stay under their root regardless of own date; due/search filters find cross-tree work. Search matches title/notes, retains ancestor context, and restores collapse state when cleared. Debounce 150 ms through typed timer generations. Losing selection selects a visible neighbor or the empty state.

| Context | Keys | Behavior |
| --- | --- | --- |
| List | `j/k`, arrows, `g/G` | Move selection and scroll |
| List | `h/l`, Left/Right | Collapse/expand branch; leaf no-op |
| Browsing | Tab/Shift+Tab | Cycle list/details; detail arrows scroll |
| Browsing | `a`, `e`, `d`, `x`/Space | Add, edit, confirm delete, toggle via R8 |
| Browsing | `/`, `r`, `?`, `q` | Search, refresh, help, quit |
| Search/form | Text, Tab, Enter, Esc | Input, focus, submit/accept search, cancel/restore focus |
| Modal | Ordinary text including `q`, `d`, `?` | Input only; no global mutation/navigation |
| Any | Ctrl+C | Cancel/exit and restore terminal |

One `d` opens confirmation; another `d` is not consent. Default button is Cancel. Forms expose title, notes, priority, due, tags and parent/root; status and manual leaf progress are edit-only, matching CLI parity and creation defaults. Field errors stay visible. Saving disables duplicate submit/modal close; Ctrl+C follows R22. Enter in notes inserts a newline; Ctrl+S or the focused Save button submits. Dirty-form cancellation defaults to Keep editing. Markdown is display-only: no remote fetch, link launch, or execution. Successful save followed by failed refresh says “saved; refresh failed”, never invites resubmission.

Feature 005 refines these interactions without changing durable service behavior:
`tusk tui` requires capable terminal stdin/stdout, rejects TERM=dumb before open,
and offers monochrome content under NO_COLOR (terminal cursor controls remain).
The filter modal exposes status/priority/tags/due and live search retains ancestry.
Root View returns a frame prepared by Update to avoid pointer-backed widget cache
writes. Unchanged raw text is preserved even when widgets cannot round-trip it;
explicit replacement is default-Cancel. Unknown outcomes require fresh-owner
readback and discarded uncertain intent before new writes. See the Feature 005
pack for exact contracts and local/native evidence boundaries.

### Rollout and gates

| Gate | Owner and closure evidence | Blocking effect |
| --- | --- | --- |
| G1 | Each feature owner synchronizes its plan/verification/workorder with this product contract | Feature 002–006 planning packs completed; planning does not authorize implementation |
| G2 | Feature 002 KTD1 and durable acceptance evidence record Go 1.25.0, modernc v1.58.0, libc v1.75.6 and SQLite 3.53.4 local proof | Local prerequisite satisfied; native/hosted release proof remains Phase 6. Feature 003 retains the installed graph |
| G3 | Phase 4/5 owners pin CLI/UI dependencies before their first unit; U15/U20 supply measurement | Feature 004/005 local graph, minimum-Go and measured CLI/TUI acceptance recorded; release-candidate/native proof remains G4 |
| G4 | Phase 6 maintainer proves exact candidate on OS/architectures and terminals, licenses, distribution destination, release metadata | Blocks publication; local green/cross-compilation are insufficient |

G2 planning choice is now owned by [Feature 002 KTD1](2026-09-06-002-feat-sqlite-storage-and-repository-plan.md#key-technical-decisions): raise the build minimum to Go 1.25 with the pinned patched driver/libc. This is a technical planning default, not a claim of explicit user approval or executed compatibility proof. The earlier affected Go 1.24 candidate is not selected. [Affected engine](https://pkg.go.dev/modernc.org/sqlite@v1.46.1), [WAL fix](https://www.sqlite.org/wal.html#walresetbug), [selected module](https://proxy.golang.org/modernc.org/sqlite/@v/v1.58.0.mod).

`sqlc` configuration `version: "2"` is a config format. Feature 002 pins the prebuilt [v1.31.1](https://github.com/sqlc-dev/sqlc/releases/tag/v1.31.1) executable and five host archive digests; its Go 1.26 source-build minimum stays outside the Go 1.25 runtime module.

First data command creates an empty database. Current installations check compatibility before migration. Later data-bearing upgrades require a consistent offline backup; down migrations only prove reversibility in test fixtures. Downgrade does not rewrite schema. Legacy PostgreSQL import needs a separate plan. Removing/replacing a binary must not delete data. Release automation must not publish without the release action being authorized.

## Implementation Units

These units are handoff contracts, not completed work or activation of future phases. Feature 001 remains the core baseline. Every new unit requires a red test before implementation, focused green evidence, applicable scenarios, and `make validate` before checklist closure. Tests excluded under `-short` must execute under full/race targets.

| Unit | Feature unit | Responsibility / primary files | Depends on |
| --- | --- | --- | --- |
| U24 | 002-6 | Pinned runtime and compatibility proof: `internal/storage/compatibility_test.go` | Feature 002 G1 and G2 planning choices; existing Feature 001 |
| U1 | 002-1 | Embedded schema and atomic migrations: `db/embed.go` | U24 compatibility proof |
| U2 | 002-2 | Reproducible sqlc queries: `db/queries.sql` | U1 |
| U3 | 002-3 | Connection and file lifecycle: `internal/storage/open.go` | U1/U2 |
| U4 | 002-4 | Repository and transaction contracts: `internal/ports/task_repository.go` | U3 |
| U5 | 002-5 | Disk concurrency and recovery proof: `internal/storage/concurrency_test.go` | U4 |
| U6 | 003-1 | Application commands and patch contracts: `internal/ports/task_service.go` | Phase 2 accepted; G1 for Feature 003 |
| U7 | 003-2 | Deterministic dates and identity: `internal/service/dateparse/` | U6 |
| U8 | 003-3 | Atomic hierarchy and rollup mutations: `internal/service/mutations.go` | U6/U7 |
| U9 | 003-4 | Queries, statistics and history orchestration: `internal/service/queries.go` | U8 |
| U10 | 003-5 | Service integration and stale-write proof: `internal/service/integration_test.go` | U9 |
| U11 | 004-1 | Lazy CLI routing and process exit: `internal/cli/root.go` | Phase 3 accepted; G1 and G3 dependency decision for Feature 004 |
| U12 | 004-2/004-7 | Scriptable mutations and confirmed deletion: `internal/cli/add.go` | U14 |
| U13 | 004-3 | List, tree, stats and history commands: `internal/cli/list.go` | U12 |
| U14 | 004-5/004-4 | Stable JSON and human formatting: `internal/cli/format.go` | U11 |
| U15 | 004-6 | Real CLI workflows and latency: `internal/cli/process_test.go` | U13 |
| U16 | 005-1/005-7/005-2 | Pure TUI model and deterministic layout: `internal/tui/model.go` | Phase 4 accepted per masterplan; G1 and G3 dependency decision for Feature 005 |
| U17 | 005-3 | Navigation, filtering and refresh generations: `internal/tui/navigation.go` | U16 |
| U18 | 005-4 | Details, Markdown and timeline: `internal/tui/details.go` | U17 |
| U19 | 005-5/005-8 | Forms and confirmed mutations: `internal/tui/forms.go` | U18 |
| U20 | 005-6 | TUI workflow and real terminal acceptance: `internal/tui/workflow_test.go` | U19 |
| U21 | 006-1 | Hosted platform quality gates: `.github/workflows/ci.yml` | Phases 4/5 accepted; G1 for Feature 006 |
| U22 | 006-2/006-6/006-5/006-7/006-8 | Release payloads, provenance, native lifecycle and authorized distribution: `.goreleaser.yaml` | U21; phase-local completion/license prerequisites; G4 decisions |
| U23 | 006-3/006-4/006-8 | Storage-free completions, public documentation/metadata and final handoff: `internal/cli/completion.go`, `README.md` | U21 for generation; U22 acceptance for final publication |

Feature 004 refines the CLI handoff order without changing product IDs: U11 → U14 → U12 → U13 → U15. Feature-local mapping is U11→U1, U14→U5/U4, U12→U2/U7, U13→U3, U15→U6. Formatting precedes its command consumers; deletion is separately provable. The Feature 004 triplet owns exact command/fixture/target details.

Feature 006 refines the release handoffs without renumbering product U21–U23:
its local order is U1 → U3 → U4 → U2 → U6 → U5 → U7 → U8. Product U21→U1,
U22→U2/U6/U5/U7/U8 and U23→U3/U4/U8. Completion/license/documentation preparation
precedes packaging; hosted candidate bytes precede native acceptance, which
precedes public release/tap/metadata activation. These umbrella handoffs do not
form a separate executable U21→U22→U23 sequence.

### U24. Pinned storage runtime and compatibility proof

- **Goal / requirements:** Prove the selected runtime before schema code; R12, R27, R29. Feature unit 002-6 / Feature 002 U6.
- **Dependencies:** Feature 002 planning pack, G2 planning choice, existing Feature 001.
- **Ownership:** `go.mod`, `go.sum`, `scripts/setup.sh`, `scripts/test/test_scripts.sh`, `Makefile`, `internal/storage/connection.go`, `internal/storage/compatibility_test.go`.
- **Approach:** Feature 002 KTD1/KTD3/KTD4 define the pinned graph and private connection factory. Prove engine version, minimum compiler, transaction mode, query-only readers and replacement pragmas before U1.
- **Red-first test:** TestSQLiteCompatibility fails on the absent driver/factory; fake setup compiler 1.24 is rejected. Record actual observations during execution.
- **Verification:** make test-compat and make build-storage (add here), make test, make race, make validate. TUSK-V01 and Feature 002 V01–V08; cross-builds do not prove native target execution.
- **Failure / recovery:** Failed compatibility blocks migration implementation; no silent pin or minimum change.
- **Reviews:** Feasibility, dependencies, concurrency, portability, evidence quality.

### U1. Embedded schema and atomic migrations

- **Goal / requirements:** Embedded schema and atomic migrations; R3–R12, R18, R29. Feature unit 002-1.
- **Dependencies:** U24 compatibility proof.
- **Ownership:** `.gitattributes`, `db/embed.go`, `db/embed_test.go`, `db/migrations/`, `internal/storage/migrations.go`, `internal/storage/migrations_test.go`, `internal/storage/schema_test.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** KTD3/KTD6/KTD7/KTD9: ledger, tasks and task_events schema; transactional up/down fixture behavior; reject foreign/newer schemas before intentional persistent changes. Feature 002 adds application identity and LF-controlled migration hashes. `db/embed_test.go` proves inventory/content rather than adding a coverage exemption.
- **Red-first test:** TestMigrate_FailurePreservesPreviousVersion and TestMigrate_NewerSchemaUnchanged: missing migrator first fails to compile; injected second-statement failure must later leave old rows/schema/version intact.
- **Verification:** make test; make race. Scenarios TUSK-V02, TUSK-V03, TUSK-V04. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Rollback the migration and retain original files; no automatic down-migration or delete/recreate on installed data.
- **Reviews:** Architecture, data integrity, migration, reliability. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U2. Reproducible sqlc queries

- **Goal / requirements:** Reproducible sqlc queries; R10–R14, R18, R29. Feature unit 002-2.
- **Dependencies:** U1.
- **Ownership:** `db/queries.sql`, `sqlc.yaml`, `internal/storage/sqlc/`, `internal/storage/queries_test.go`, `scripts/sqlc.sh`, `Makefile`, `scripts/test/test_scripts.sh`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** KTD3/KTD8: pin generator/checksum, compile SQLite queries, isolate generated output; add canonical generate/check-generated targets with script tests. Include transactional row/event and subtree/ancestor queries.
- **Red-first test:** TestQueries_FilterAndSubtreeParity fails on missing queries; malformed SQL fixture must make generation fail; stale generated output must fail check-generated.
- **Verification:** make test; make setup-sqlc, make generate, make check-generated and make test-scripts (add here). Scenarios TUSK-V05, TUSK-V06, TUSK-V07. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Generation failure cannot overwrite checked-in output with partial files; preserve source and regenerate deterministically.
- **Reviews:** API contract, SQL correctness, performance, maintainability. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U3. Connection and file lifecycle

- **Goal / requirements:** Connection and file lifecycle; R1, R2, R11, R12, R23, R29. Feature unit 002-3.
- **Dependencies:** U1/U2.
- **Ownership:** `internal/storage/open.go`, `internal/storage/path.go`, `internal/storage/open_test.go`, `internal/storage/path_test.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** KTD4/KTD9: one writer/up to four readers, per-connection pragmas, lazy path opening, safe permissions, close ownership, disk and uniquely named memory fixture factories.
- **Red-first test:** TestOpen_PathPrecedenceAndLiteralFilename, TestOpen_ReplacementConnectionPragmas: absent opener fails; later failures expose ignored XDG, unsafe DSN, or unconfigured replacement connection.
- **Verification:** make test; make race. Scenarios TUSK-V08, TUSK-V09, TUSK-V10, TUSK-V11. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Close partially opened handles on error; never switch to another path or remove DB/sidecars. Failed close returns context without masking the primary failure.
- **Reviews:** Security/privacy, lifecycle, concurrency, portability. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U4. Repository and transaction contracts

- **Goal / requirements:** Repository and transaction contracts; R3–R7, R10, R13, R14, R18, R29. Feature unit 002-4.
- **Dependencies:** U3.
- **Ownership:** `internal/ports/task_repository.go`, `internal/ports/errors.go`, `internal/storage/sqlite_repository.go`, `internal/storage/sqlite_repository_test.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** KTD5/KTD6/KTD8: explicit read/write callback ports, strict row conversion, bind-only SQL, nested transaction rejection, graph/metadata retrieval and error mapping.
- **Red-first test:** TestRepository_RoundTripAndDetachedValues and TestWithWrite_ChildAndHistoryRollback fail on missing methods; invalid decoded rows, missing IDs and callback error must never return partial success.
- **Verification:** make test; make race; make validate. Scenarios TUSK-V12, TUSK-V13, TUSK-V14, TUSK-V15. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Rollback every failed callback; discard poisoned connection; preserve errors.Is/context errors. Returned values must not alias shared repository state.
- **Reviews:** Contract, correctness, data integrity, simplicity. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U5. Disk concurrency and recovery proof

- **Goal / requirements:** Disk concurrency and recovery proof; R10–R12, R28, R29. Feature unit 002-5.
- **Dependencies:** U4.
- **Ownership:** `internal/storage/concurrency_test.go`, `internal/storage/recovery_test.go`, `internal/storage/storage_bench_test.go`, `internal/storage/testdata/`, `docs/storage.md`, `Makefile`, `phase 002 triplet`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Use deterministic lock barriers, two independent process connections, temporary disk files, crash checkpoints and SQLite integrity checks; memory tests never stand in for WAL evidence.
- **Red-first test:** TestConcurrentWriters_NoLostUpdate and TestRecovery_KilledWriterAtomic expose incomplete isolation/recovery. If existing behavior already passes, record characterization; introduce no implementation change without a distinct failing case.
- **Verification:** make test; make race; make validate. Scenarios TUSK-V16, TUSK-V17, TUSK-V18, TUSK-V19. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Crash before commit leaves old state; after acknowledged commit and process termination read committed state. Keep locks/readers bounded and prove another operation works after cancellation.
- **Reviews:** Adversarial data integrity, concurrency, reliability, performance. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U6. Application commands and patch contracts

- **Goal / requirements:** Application commands and patch contracts; R3–R10, R19, R20, R22, R29. Feature unit 003-1.
- **Dependencies:** Phase 2 accepted; G1 for Feature 003.
- **Ownership:** `internal/ports/task_service.go`, `internal/ports/task_commands.go`, `internal/service/task_service.go`, `internal/service/task_service_test.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Define create, atomic patch, complete, reopen, delete, list, tree, stats and history operations; inject repo, clock, ID source, zone, completion policy. Distinguish unchanged/set/clear and preserve no-op semantics.
- **Red-first test:** TestService_InvalidCommandDoesNotBeginWrite and TestService_PatchPresence fail on missing service then catch clear/omitted conflation.
- **Verification:** make test-unit; make test. Scenarios TUSK-V20, TUSK-V21. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Reject invalid pure input before opening transaction; propagate I/O failure without mutating caller objects or losing draft data.
- **Reviews:** API contract, product parity, correctness, simplicity. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U7. Deterministic dates and identity

- **Goal / requirements:** Deterministic dates and identity; R3, R4, R15, R16, R29. Feature unit 003-2.
- **Dependencies:** U6.
- **Ownership:** `internal/service/dateparse/`, `internal/service/id.go`, `internal/service/id_test.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** KTD11 and R15/R16; table-driven calendar parsing and UUIDv7 bit/layout validation with injected entropy/time. Include zone/DST/month-end fixtures.
- **Red-first test:** TestParseDue_CalendarDSTAndInvalid, TestUUIDv7_EntropyFailure: missing helpers fail; elapsed-hour date arithmetic must fail DST cases and entropy failure must return no ID.
- **Verification:** make test-unit. Scenarios TUSK-V22, TUSK-V23, TUSK-V24, TUSK-V25. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Invalid dates/overflow/entropy abort before write; never substitute now, random fallback, or partial normalized values.
- **Reviews:** Correctness, portability, API compatibility. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U8. Atomic hierarchy and rollup mutations

- **Goal / requirements:** Atomic hierarchy and rollup mutations; R5–R10, R18, R20, R29. Feature unit 003-3.
- **Dependencies:** U6/U7.
- **Ownership:** `internal/service/mutations.go`, `internal/service/rollup.go`, `internal/service/mutations_test.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Run all mutation graph reads inside one write callback; calculate both move chains deepest first; enforce R7–R9 and append events only for actual changes.
- **Red-first test:** TestMove_RecalculatesBothAncestorChains, TestComplete_SubtreeAndEventsAtomic, TestReopen_OverridesAutoCompletion fail when propagation, rollback or explicit-intent precedence is missing.
- **Verification:** make test-unit; make test; make race. Scenarios TUSK-V26, TUSK-V27, TUSK-V28, TUSK-V29, TUSK-V30, TUSK-V31. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Any child/ancestor/event write failure rolls back everything. Concurrent cycle-creating moves serialize so at most one succeeds.
- **Reviews:** Correctness, concurrency, data integrity, adversarial. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U9. Queries, statistics and history orchestration

- **Goal / requirements:** Queries, statistics and history orchestration; R13–R18, R29. Feature unit 003-4.
- **Dependencies:** U8.
- **Ownership:** `internal/service/queries.go`, `internal/service/queries_test.go`, `internal/service/history.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Use committed snapshots, core filter/sort parity, local-day selection and R17 metrics. Build selected subtree from validated full data without persisting root projection changes.
- **Red-first test:** TestList_UnicodeAndDayBoundaryParity, TestStats_ReopenAndDelete, TestTree_FilteredAncestorContext fail on mismatched interval, retained completion or orphaned projections.
- **Verification:** make test-unit; make test. Scenarios TUSK-V32, TUSK-V33, TUSK-V34, TUSK-V35. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Return errors on corrupt graph/history; no silent omission, partial metrics, or fabricated timeline from updated_at.
- **Reviews:** Query correctness, contract, performance. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U10. Service integration and stale-write proof

- **Goal / requirements:** Service integration and stale-write proof; R6–R10, R18–R20, R22, R29. Feature unit 003-5.
- **Dependencies:** U9.
- **Ownership:** `internal/service/integration_test.go`, `internal/service/conflict_test.go`, `phase 003 triplet`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Replay lifecycle fixtures against real SQLite plus deterministic fake failures; prove stale draft/delete confirmation behavior and error recovery across process boundaries.
- **Red-first test:** TestStaleForm_ConflictPreservesNewerEdit and TestDelete_ChangedConfirmedSubtree expose missing authoritative rechecks; retain red evidence for any required fix.
- **Verification:** make test; make race; make validate. Scenarios TUSK-V36, TUSK-V37, TUSK-V38. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Conflict preserves prior committed data and user intent; reload before retry. Write acknowledgment and read-refresh failure are distinct outcomes.
- **Reviews:** Adversarial correctness, concurrency, error propagation. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U11. Lazy CLI routing and process exit

- **Goal / requirements:** Lazy CLI routing and process exit; R1, R2, R9, R19, R22, R23, R29. Feature unit 004-1.
- **Dependencies:** Phase 3 accepted; G1 and G3 dependency decision for Feature 004. U15 closes G3's later measurement portion.
- **Ownership:** `internal/cli/root.go`, `internal/cli/root_test.go`, `cmd/tusk/main.go`, `cmd/tusk/main_test.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Parse Cobra tree/arity/help before resolving data configuration; inject streams and a lazy service factory. Preserve immutable version behavior and help on empty invocation.
- **Red-first test:** TestRoot_SyntaxAndHelpNeverOpenStorage, TestRoot_ExitCodes fail against scaffold's unknown-command success and expose eager storage initialization.
- **Verification:** make test-unit; make test. Scenarios TUSK-V39, TUSK-V40, TUSK-V41. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Usage only for syntax errors; operational errors do not dump usage or leak user notes. Close resources at command end and respect cancellation.
- **Reviews:** CLI contract, startup performance, reliability. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U12. Scriptable mutation commands

- **Goal / requirements:** Scriptable mutation commands; R3–R9, R18–R23. Feature units 004-2 and 004-7.
- **Dependencies:** U14 (feature U5/U4 formatting/output), then feature U2 before U7.
- **Ownership:** `internal/cli/add.go`, `internal/cli/edit.go`, `internal/cli/done.go`, `internal/cli/delete.go`, `internal/cli/mutations_test.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Bind explicit patch/clear fields to service commands; no UI-owned domain logic. Implement recursive and force distinction, default-No prompt, policy flag and validation mapping.
- **Red-first test:** TestDelete_ForceDoesNotImplyRecursive, TestEdit_ClearVersusOmitted, TestDone_Descendants fail on missing flags/incorrect service mapping.
- **Verification:** make test-unit; make test. Scenarios TUSK-V42, TUSK-V43, TUSK-V44. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Never prompt on JSON/non-TTY; no hidden destructive retries or partial output. Confirm conflicts return actionable error.
- **Reviews:** CLI ergonomics, destructive action boundary, agent parity. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U13. List, tree, stats and history commands

- **Goal / requirements:** List, tree, stats and history commands; R13–R18, R19, R21–R23. Feature unit 004-3.
- **Dependencies:** U12.
- **Ownership:** `internal/cli/list.go`, `internal/cli/tree.go`, `internal/cli/stats.go`, `internal/cli/history.go`, `internal/cli/queries_test.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Map R13–R18 query contracts exactly; include full IDs and timeline access. Filters never implicitly change persisted state.
- **Red-first test:** TestQueries_EmptyAndMissingRoot, TestHistory_ParityWithDetails fail on missing routes or null/empty inconsistencies.
- **Verification:** make test-unit; make test. Scenarios TUSK-V45, TUSK-V46. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Corruption/query failure emits no fabricated empty-success response; no diagnostics on stdout.
- **Reviews:** API contract, product completeness, query correctness. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U14. Stable human and JSON formatting

- **Goal / requirements:** Stable human and JSON formatting; R14, R19, R21–R23, R28. Feature unit 004-4/004-5.
- **Dependencies:** U11; execute feature U5 JSON/output, then U4 human formatting, before command consumers.
- **Ownership:** `internal/cli/format.go`, `internal/cli/json.go`, `internal/cli/format_test.go`, `internal/cli/testdata/`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Explicit Task/tree/stats/event DTOs, one JSON value+LF, tabular colors only under R23, cell-aware truncation and terminal-control sanitization. Golden files test deliberate format, not incidental styling.
- **Red-first test:** TestJSON_ExplicitNullsAndArrays, TestFormat_ControlSequencesAndUnicode, TestOutputFailure_NoMutationRetry fail on core omitempty, unsafe control text or ignored writer errors.
- **Verification:** make test-unit; make test. Scenarios TUSK-V47, TUSK-V48, TUSK-V49. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Encoding before output may fail cleanly; a broken output stream may be partial but never contains injected human text or triggers replay.
- **Reviews:** Serialization, security, accessibility, compatibility. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U15. Real CLI workflows and latency

- **Goal / requirements:** Real CLI workflows and latency; R1, R2, R5–R10, R19–R23, R28, R29. Feature unit 004-6.
- **Dependencies:** U13 and all earlier Feature 004 units.
- **Ownership:** `internal/cli/process_test.go`, `scripts/cli-bench/`, `scripts/test/test_scripts.sh`, `Makefile`, `phase 004 triplet`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Feature U1 adds test-cli/build-cli; U6 adds bench-cli; build once via make build, run real subprocesses against temp homes/files. Measure launch through exit with encoded output consumed, not go run or core microbenchmarks.
- **Red-first test:** TestProcess_CaptureCompleteReopenDelete and TestHelp_NoFilesystemEffects fail against any adapter integration regressions; latency assertions start only with a documented stable fixture.
- **Verification:** make test; make race; make test-cli (added by feature U1) and make bench-cli (add here); make validate. Scenarios TUSK-V50, TUSK-V51, TUSK-V52, TUSK-V53. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Retain raw samples/outliers and exact binary revision; slow/busy runs are not silently discarded. Query committed state after interrupted mutations.
- **Reviews:** End-to-end, performance, non-regression. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U16. Pure TUI model and deterministic layout

- **Goal / requirements:** Pure TUI model and deterministic layout; R23–R26, R29. Feature units 005-1/005-7/005-2.
- **Dependencies:** Phase 4 accepted per masterplan; G1 and G3 dependency decision for Feature 005. U20 closes G3's later measurement portion.
- **Ownership:** `internal/tui/model.go`, `internal/tui/layout.go`, `internal/tui/model_test.go`, `internal/tui/layout_test.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** KTD13/R24–R26: constructor-initialized components, typed commands, layout in Update, bounded small-terminal state and content-independent geometry.
- **Red-first test:** TestView_RepeatedCallsDeepStateUnchanged and TestResize_CellBounds fail on missing model or any slice/style mutation and overflow.
- **Verification:** make test-unit; make test. Scenarios TUSK-V54, TUSK-V55. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Load/resize errors show a bounded visible state; no panic, negative dimensions, hidden draft loss, or View-triggered work.
- **Reviews:** UI correctness, purity, layout, accessibility. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U17. Navigation, filtering and refresh generations

- **Goal / requirements:** Navigation, filtering and refresh generations; R13, R14, R24–R26. Feature unit 005-3.
- **Dependencies:** U16.
- **Ownership:** `internal/tui/navigation.go`, `internal/tui/messages.go`, `internal/tui/navigation_test.go`, `internal/tui/refresh_test.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** KTD14, TUI grouping/key table, selection by ID, collapsed ancestor context, typed debounce/periodic refresh, stale-read invalidation.
- **Red-first test:** TestRefresh_OldResultCannotOverwriteNewerSnapshot, TestSearch_RestoresSelectionAndCollapse fail under reversed result order and shrinking results.
- **Verification:** make test-unit; make race. Scenarios TUSK-V56, TUSK-V57, TUSK-V58. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Failed refresh preserves labeled stale snapshot; no reply may target a newer modal; all key paths remain usable on empty data.
- **Reviews:** Async races, interaction completeness, accessibility. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U18. Details, Markdown and timeline

- **Goal / requirements:** Details, Markdown and timeline; R7, R15, R18, R23–R26. Feature unit 005-4.
- **Dependencies:** U17.
- **Ownership:** `internal/tui/details.go`, `internal/tui/details_test.go`, `internal/tui/testdata/`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Scroll bounded Markdown/details/history; UTC storage/local display; sanitize terminal controls without changing persisted text. No network or link execution.
- **Red-first test:** TestDetails_LongUnicodeAndUntrustedMarkdown, TestTimeline_RefreshShowsCommittedEvents fail on overflow, hidden network behavior or fabricated history.
- **Verification:** make test-unit; make test. Scenarios TUSK-V59, TUSK-V60. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Missing selection clears details; render errors show safe plain text; last good task cannot be mislabeled as another task.
- **Reviews:** Security/privacy, UI usability, parity. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U19. Forms and confirmed mutations

- **Goal / requirements:** Forms and confirmed mutations; R4–R10, R18–R20, R22, R24–R26. Feature units 005-5/005-8.
- **Dependencies:** U18.
- **Ownership:** `internal/tui/forms.go`, `internal/tui/forms_test.go`, `internal/tui/confirm.go`, `internal/tui/confirm_test.go`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Fields cover R19 parity; focus trapped in modal; draft/base snapshot tracked; no duplicate submissions; one-key delete confirmation defaults Cancel.
- **Red-first test:** TestForm_QDoesNotQuitAndFailureKeepsDraft, TestSave_RefreshFailureDoesNotResubmit, TestConfirm_ChangedSubtreeConflict fail on missing focus/commit tracking.
- **Verification:** make test-unit; make test; make race. Scenarios TUSK-V61, TUSK-V62, TUSK-V63. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Validation/conflict/failure retain draft; successful commit closes saved form even if refresh fails; uncertain commit requires reload.
- **Reviews:** Race conditions, destructive action UX, correctness, parity. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U20. TUI workflow and real terminal acceptance

- **Goal / requirements:** TUI workflow and real terminal acceptance; R23–R26, R28, R29. Feature unit 005-6.
- **Dependencies:** U19.
- **Ownership:** `internal/tui/workflow_test.go`, `internal/tui/bench_test.go`, `Makefile`, `phase 005 triplet`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Synthetic message sequences cover normal/empty/error/recovery states. Add canonical bench-tui. Separately record real PTY/terminal resize, keyboard, plain-mode readability and cleanup.
- **Red-first test:** TestWorkflow_LoadEditConflictRecoverAndQuit exposes sequence defects; pure View benchmark is characterization until a measured regression produces a red test.
- **Verification:** make test; make race; make bench-tui (add here); make validate; manual terminal matrix. Scenarios TUSK-V64, TUSK-V65, TUSK-V66. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Cancelled/failed launch restores terminal; rerun app afterward. Headless green cannot close real-terminal acceptance.
- **Reviews:** End-to-end UI, accessibility, lifecycle, performance. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U21. Hosted platform quality gates

- **Goal / requirements:** Hosted platform quality gates; R1, R12, R27–R29. Feature unit 006-1.
- **Dependencies:** Phases 4/5 accepted; G1 for Feature 006.
- **Ownership:** `.github/workflows/ci.yml`, `scripts/setup.sh`, `scripts/test/test_scripts.sh`, `Makefile`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Feature 006 U1 pins Actions/tools, executes five native runner targets plus Go 1.25.0 compatibility, declares Windows Bash/Make/native race compiler prerequisites and fails formatting/generated/module drift. Its triplet owns exact pins, commands and candidate-SHA evidence.
- **Red-first test:** Script fixture with absent compiler/wrong generator must fail the intended prerequisite; workflow validation must reject missing platform/race gates before final configuration.
- **Verification:** make setup; make validate; hosted matrix on candidate SHA. Scenarios TUSK-V67, TUSK-V68. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Missing architecture/runtime proof remains pending, not silently skipped; least-privilege read-only PR jobs do not publish artifacts as a release.
- **Reviews:** Portability, deployment, supply chain, evidence quality. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U22. Release artifacts and installation lifecycle

- **Goal / requirements:** Release artifacts and installation lifecycle; R1, R2, R11, R12, R27–R30. Feature units 006-2/006-6/006-5/006-7/006-8.
- **Dependencies:** U21; phase-local completion/license preparation; G4 owner decisions and native/hosted acceptance before publication.
- **Ownership:** `.goreleaser.yaml`, `Makefile`, `scripts/release_check.sh`, `scripts/test/test_scripts.sh`, `release documentation`. Test paths are planned unless already present; do not overwrite another unit's files without coordination.
- **Approach:** Feature 006 packages immutable version constants via a build overlay, all five CGO-free payloads, source/notice/checksum manifests and trusted provenance. Hosted candidates are tested and promoted as the same bytes. Prove native backup/upgrade/refusal/remove preservation and current macOS Homebrew cask delivery; owner license/destination/version/publication remain explicit G4 gates.
- **Red-first test:** Release manifest fixture missing architecture/checksum/license must fail; existing-data upgrade fixture must detect recreation or destructive downgrade.
- **Verification:** make release-check and make release-snapshot (add here); hosted artifact execution; make validate. Scenarios TUSK-V69, TUSK-V70, TUSK-V71. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Snapshot never publishes; failed upgrade preserves original DB; no unsigned/unverified artifact is represented as published/accepted.
- **Reviews:** Deployment, data lifecycle, portability, supply chain. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

### U23. Completions, documentation and final handoff

- **Goal / requirements:** Completions, documentation and final handoff; R1, R19, R27, R29, R30. Feature units 006-3/006-4/006-8.
- **Dependencies:** U21 for generation; U22 native/release acceptance for final public activation. Follow the phase-local sequence above.
- **Ownership:** `internal/cli/completion.go`, `internal/cli/completion_test.go`, `scripts/test/test_completions.sh`, `Makefile`, `README.md`, `docs/install.md`, `docs/releasing.md`, community/metadata content and the phase 006 triplet. Planned files remain unimplemented.
- **Approach:** Static Bash/Zsh/Fish completion and deterministic man pages reuse the fresh Cobra tree without storage. Feature 006 U4/U8 give README/install/use/backup/security/license instructions, sanitized evidence assets and concrete GitHub About/topics/homepage/social preview, then inspect actual rendered/readback state after authorized publication.
- **Red-first test:** TestCompletion_NoStorageWithBrokenPath and shell parse fixtures fail for eager initialization or malformed output; documentation examples must reproduce declared CLI behavior.
- **Verification:** make test; make test-completions (add here); make validate; G4 evidence review. Scenarios TUSK-V72, TUSK-V73. Record actual red/green evidence during execution; these are expected failures, not observed results.
- **Failure / recovery:** Missing shell/host evidence is recorded pending; neither docs nor master checklist claims publication before release actually occurs.
- **Reviews:** Documentation usability, CLI compatibility, release evidence. Close the unit only with the applicable aggregate gate and phase triplet/master checklist update.

## Verification Contract

The [verification plan](../verification-plans/2026-09-06-001-feat-tusk-modern-task-system-verification-plan.md) maps requirements to units/scenarios. The [workorder](../workorders/2026-09-06-001-feat-tusk-modern-task-system-issues-workorder.md) separates planning resolutions from open gates/runtime defects.

Existing commands: `make setup`, `make test-unit`, `make test`, `make race`, `make validate`, `make build`, `make coverage`, `make bench-tree`, `make bench-build`. The Makefile has no TEST/PKG/RUN filtering variables; do not pass ignored variables as supposed focused proof. Run the canonical suite containing the named red test until filtering support is deliberately added and tested.

Planned additions, not available now: `make test-compat`, `make build-storage`, `make setup-sqlc`, `make generate`, `make check-generated`, `make test-scripts`, `make bench-storage`, `make test-cli`, `make bench-cli`, `make bench-tui`, `make release-check`, `make release-snapshot`, `make test-completions`. Ownership is in the units. Current build is native-only. Never expand `make clean` to remove an installed database.

This planning pass inspects documents/source/manifests and audits links, IDs, tables, status, and whitespace. It does not run application tests or install dependencies. Execution records later require date, exact revision, environment, command, red/green result, and evidence location. Test-file presence is not a passing scenario.

## Definition of Done

Planning is complete when the triplet is coherent, contracts are preserved or explicitly reconciled, all units have dependencies/ownership/red tests/failure behavior/reviews, all requirements map to scenarios, and every gate has an owner/action/closure condition. Product planning completion does not close G1 or make this umbrella executable.

Implementation completion requires local scenarios, canonical gates, hosted checks, and manual terminal/lifecycle acceptance on the candidate revision. Remove abandoned experiments, settle issues, update the phase triplet and master checklist. Implementation/release boxes remain unchecked until evidence exists. Publication is separate from local acceptance.

## Sources and Research

Repository: [task lifecycle](../../internal/core/task.go), [tree](../../internal/core/tree.go), [rollup](../../internal/core/rollup.go), [filter/sort](../../internal/core/filter.go), [errors](../../internal/core/errors.go), [Makefile](../../Makefile), [concepts](../../CONCEPTS.md), [feature registry](README.md), [research baseline](../research/modern-stack-2026.md).

External documentation checked 2026-09-08 is cited beside the decisions. Module inspection is planning evidence; no dependency graph was installed/executed. Recheck pins at G2/G3. Path rules follow [XDG](https://specifications.freedesktop.org/basedir/latest/); Tusk's home fallback is a cross-platform product choice, not a claim every OS uses XDG natively.

## Feature 002 local handoff — 2026-09-08

The [Feature 002 triplet](../plans/2026-09-06-002-feat-sqlite-storage-and-repository-plan.md) now records six implemented units with per-unit validated commits and 67/68 local scenarios accepted. [Storage operations and Phase 3 obligations](../storage.md) and [durable evidence](../verification-evidence/002/README.md) cover the repository boundary, atomic metadata history, migration refusal, process recovery and benchmarks. Product U24 compatibility evidence is available locally; target-native TUSK-V01 acceptance remains pending. This handoff does not check cross-phase service/CLI/TUI or hosted/native product scenarios. At that handoff MASTERPLAN.md advanced to Feature 003 planning. The subsequent Feature 003 handoff below supersedes its former outline status.

## Feature 003 planning handoff — 2026-09-09

The [service plan](2026-09-06-003-feat-task-service-engine-plan.md), [verification plan](../verification-plans/2026-09-06-003-feat-task-service-engine-verification-plan.md), and [workorder](../workorders/2026-09-06-003-feat-task-service-engine-issues-workorder.md) complete G1 for Feature 003: 22 feature requirements, seven bounded units, 91 unexecuted scenarios and 16 planning findings/gates. Product requirements and the 24 product handoff IDs remain unchanged. Feature 003 specifies internal API types, no-op/base/consent equality, date/DST/UUID rules, transaction outcome propagation, deterministic net-change history and real disk service acceptance.

| Product handoff | Feature 003 execution mapping |
| --- | --- |
| U6 / 003-1 | Feature U1 contracts, constructor/validation seams and safe-error compatibility |
| U7 / 003-2 | Feature U2 calendar and UUIDv7 helpers |
| U8 / 003-3 | Feature U3 rollup/change-set helpers plus new U6 creation/patch and U7 lifecycle/deletion, split from original 003-4 |
| U9 / 003-4 | Feature U4 snapshot queries and completed service facade |
| U10 / 003-5 | Feature U5 disk/concurrency/recovery, compatibility, performance and consumer handoff |

Current source inspection: clean main `679f5cefd6e626a3c67c1e08a433ab786c944883`, Go 1.25.0, modernc v1.58.0/libc v1.75.6, core/ports/storage present, service absent. Feature 002's recorded local acceptance supplies the storage prerequisite; its native/hosted limitations remain. Feature 003's U1 → U2 → U3 → U6 → U7 → U4 → U5 ordering supersedes the old five-unit outline while preserving product handoff IDs. Only planning checkboxes close; no runtime test, code change or implementation authorization is claimed.

Consumer handoffs: Feature 004 must embed production timezone data, map structured commands/base/consent and typed outcomes to the existing CLI grammar/DTOs, and prove process latency. Feature 005 owns draft/refresh/terminal acceptance; Feature 006 owns native/hosted releases. These remain their respective G1/G3/G4 obligations.

## Feature 004 execution handoff — 2026-09-28

Feature 004 has six separately validated unit commits; U6 remains active. Its
[verification matrix](../verification-plans/2026-09-06-004-feat-cli-interface-and-scripting-verification-plan.md)
checks V01–V86 and V88–V89 with [local evidence](../verification-evidence/004/README.md).
Canonical tests, race, coverage, generated checks, minimum Go and five-target
cross-builds pass. Actual Linux process/PTY recovery and visible Kitty checks pass.
V87 reference latency fails on the 1,000-task fixture; ISS-023 blocks U6/Phase 4
acceptance and Phase 5 advancement. No performance contract was weakened.

[CLI](../cli.md) and [service](../service.md) document the current consumer contract.
Feature 005 owns TUI registration, drafts, consent and refresh using the same ports
and explicit outcome handling. Feature 006 owns completion/man pages, native
Windows/macOS and architecture runtime, hosted checks and release artifacts. CLI
V90–V91 and carried storage/service native obligations remain pending. This handoff
does not close cross-phase product scenarios or authorize publication.

### Feature 004 broader optimization follow-up — 2026-09-28

The owner retained 15 ms. Storage/query allocation and sorting changes pass
canonical validation, minimum-Go tests/cross-builds and fresh Kitty checks, but
the all-sample latency gate still fails. Feature 004 U6 and Phase 4 stay open;
no downstream implementation or acceptance is inferred. See the synchronized
[Feature 004 evidence](../verification-evidence/004/u6-broader-checkpoint.json).

## U6 local acceptance — 2026-09-28

Feature 004 U6 and Phase 4 are locally accepted under the owner-delegated distribution
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

## Feature 005 planning handoff — 2026-09-29

The [Feature 005 plan](2026-09-06-005-feat-interactive-tui-application-plan.md), [verification matrix](../verification-plans/2026-09-06-005-feat-interactive-tui-application-verification-plan.md) and
[workorder](../workorders/2026-09-06-005-feat-interactive-tui-application-issues-workorder.md) complete G1 for TUI planning: 28 feature requirements, eight
units, 112 unexecuted scenarios and 25 findings/gates (20 corrected in planning,
five execution/release gates). Product U16→feature U1/U7/U2, U17→U3, U18→U4,
U19→U5/U8 and U20→U6. Order: U1 → U7 → U2 → U3 → U4 → U5 → U8 → U6.
Existing product requirement, handoff and scenario IDs/check states are preserved.

Local main at e899491 records Feature 004 PR #4 merged; its local acceptance
receipts remain valid historical evidence, not rerun by this pass. Feature 005
selects its v1 dependency graph and defines prepared-frame purity, single-operation
service ownership, draft/consent conflicts, raw-text preservation and fresh-owner
unknown-outcome readback. G3 remains open for its U1 build proof/U6 measurement;
V01–V110 are local feature obligations, V111–V112 stay with Feature 006 native and
hosted release proof. The TUI source/Makefile targets do not exist yet.

TUSK-V54–V66 now map to that detailed matrix, with TUSK-V35/V38 and governance
covered at the consumer boundary. No product scenario or implementation/release
checkbox closes from planning. No application tests, make validate, benchmarks
or Kitty acceptance ran. MASTERPLAN's next unit is Feature 005 U1 only after
an implementation instruction. Historical handoffs above retain their dated
meaning; this is the current planning handoff.

### Feature 005 runtime compatibility correction — 2026-09-29

Feature 005 U7 exposed an eager global terminal-color query in Bubble Tea
v1.3.10's package initializer. The local dependency copy removes only that
initializer, retains the upstream version/APIs/license, and records source
hashes in `third_party/bubbletea/TUSK-PATCH.json`. Actual child PTY regressions
prove the unsolicited query is absent and ordinary CLI/TUI startup works again.
This amends the selected v1 runtime without a v2 migration or weaker CLI latency
gate. U6 still owns final-candidate distributions; native/hosted proof stays with
Feature 006. See the Feature 005 U7 receipt for execution evidence.

### Feature 005 presentation checkpoint — 2026-09-29

Feature U1/U7/U2 now provide product U16's pure root, owned lifecycle and bounded
layout foundation. The owner selected spacious title/metadata rows after a
Kitty prototype and requested uniform modal/panel backgrounds. The production
surface and actual terminal resize fixes pass canonical, minimum-Go and owned
Kitty checks; see [U2 evidence](../verification-evidence/005/u2.json).
Product TUSK-V54/V55 retain their later form/selection/saving repetitions;
U17–U20 remain pending. Feature U3 is the next execution target. No release,
native-platform or final CLI-distribution acceptance is added by this checkpoint.

### Product U17 navigation handoff — 2026-09-29

Feature 005 U3 supplies forest grouping, navigation, selection identity, literal
search, filters and single-operation refresh; see the synchronized
[U3 receipt](../verification-evidence/005/u3.json). TUSK-V58 passes injected
search/timer, stale snapshot, retry and resize checks. TUSK-V56 has list/filter
and Kitty proof but retains final detail/form keyboard coverage in U4/U5/U6;
TUSK-V57 retains post-mutation/form identity proof in U5/U8. U18 is the current
feature target. No native, hosted, release or complete TUI claim is added.

### Product U18 details handoff — 2026-09-29

Feature 005 U4 supplies wrapped metadata, safe asynchronous Markdown and actual
ordered history; TUSK-V59/TUSK-V60 pass current synthetic, disk and owned Kitty
checks at all target sizes. See the [U4 receipt](../verification-evidence/005/u4.json).
Large/pathological notes use labeled full-content plain text under the explicit
formatting budget; raw storage is preserved. U19 is next. Form identity, writes,
consent/recovery UI and final combined-flow proof remain their designated units.


### Feature 005 U5 handoff — 2026-09-29

Feature 005 U5 now supplies create/edit/move/lifecycle forms with detached Base,
raw preservation, explicit conflict/discard and one-write/readback barriers.
[U5 evidence](../verification-evidence/005/u5.json) records canonical/minimum-Go
and real Kitty checks. TUSK-V62 closes. TUSK-V61/V63 keep their remaining delete
consent portions open for U8. Product U19 and Feature 005 overall acceptance remain
open through U8/U6; native/hosted release evidence remains Feature 006.


### Feature 005 U8 handoff — 2026-09-29

[U8 evidence](../verification-evidence/005/u8.json) closes product TUSK-V61/V63
and the U19 forms/consent implementation. Exact preview consent, fresh-owner
unknown-outcome readback, no replay, explicit acknowledgment and real Kitty
checks pass. U20 / Feature 005 U6 final workflow, performance, documentation and
review remain active; native/hosted release evidence remains Feature 006.

### Feature 005 U6 visual and validation checkpoint — 2026-09-29

The owner approved the refined Spacious production TUI, including padding,
tabs, checkboxes and opaque dialog backgrounds. Feature V101–V105 and V109
pass: workflow, disk/lifecycle, child PTY, owned Kitty, canonical validation,
minimum-Go/five-target builds and documentation. The visual proof is separate
from automated test output.

Product U20 remains open. The final linked CLI matrix passes 74/84 cases after
removing eager syntax/CSS registry initialization from the pinned Markdown
renderer. All failed samples are retained. Final performance and code-review
settlement are required before Phase 5 acceptance and the U6 local commit.
Native macOS/Windows and hosted release proof remain Feature 006 obligations.

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
