# Tusk CLI

Build with `make build`. The executable is `bin/tusk`. Commands use an embedded
SQLite database; no daemon or network service is required. `tusk`, `tusk help`,
`tusk --help`, `tusk -h`, `tusk version`, `tusk --version` and `tusk -v` do not
open storage. Command help is available as `tusk help add` or `tusk add --help`.

The executable defaults to one Go execution processor for its single serial
operation; an explicit `GOMAXPROCS` setting is respected. Embedded `cli.Run`
callers keep their own runtime policy. This setting does not prevent signal
handling, asynchronous cancellation, or separate CLI processes from running.

## Commands

Every data command supports `--json`. Pass complete task IDs; prefixes do not
identify tasks. Quote titles, notes, paths and expressions as one shell argument.
Use `--` before positional values that start with a dash:

```sh
tusk add --json -- '- investigate a warning'
tusk add 'Review release' --notes 'Check the changelog' --due tomorrow --tags work
tusk list --all --json
tusk edit TASK_ID --notes '' --clear-due --json
tusk done TASK_ID --json
tusk delete TASK_ID --recursive --force --json
```

Replace `TASK_ID` with the full ID from a successful response.

| Command | Flags and behavior |
| --- | --- |
| `add <title>` | `--notes/-n`, `--priority/-p`, `--due/-d`, `--parent`, `--tags/-t`. Creates a todo task; omitted priority is medium. |
| `edit <id>` | `--title`, `--notes/-n`, `--priority/-p`, `--status/-s`, `--progress`, `--due/-d`, `--clear-due`, `--parent`, `--root`, `--tags/-t`, `--clear-tags`. Only supplied fields change; at least one effective intent is required. |
| `done <id>` | Completes the target and every descendant atomically. |
| `delete <id>` | `--recursive` permits descendant deletion. `--force` skips consent. Each flag is independent. |
| `list` | `--status/-s`, `--priority/-p`, `--tags/-t`, `--due`, `--search`, `--parent`, `--root`, `--all`. Defaults to open tasks. |
| `tree [id]` | Whole forest or selected subtree, including done tasks. |
| `stats` | Counts all retained tasks, including parents. |
| `history <id>` | Metadata events in sequence order; no previous/current field values. |

Scalar flags use their last supplied value. Repeat `--status` and `--priority`
to OR values within that filter. Different filters are ANDed. Tags are repeatable
and comma-separated; list requires every supplied tag. Tags are trimmed,
normalized to lowercase, deduplicated and sorted. Search is a literal
case-insensitive title/notes substring, including `%` and `_`.

Statuses are `todo`, `in-progress`, `blocked`, `done`. Priorities are
`low/1`, `medium/2`, `high/3`, `urgent/4`. Values are case-insensitive.
Lists sort by priority descending, due ascending (missing last), creation time
ascending and ID ascending. `--all` conflicts with `--status`; `--root`
conflicts with `--parent`.

An edit is one atomic patch. Empty notes clear notes; `--clear-tags`, `--clear-due`
and `--root` clear tags, due date and parent respectively. Each clear flag
conflicts with its value flag. `--progress` accepts signed decimal integer 0–100 for leaves
(for example, `010` means 10; hexadecimal and underscores are rejected);
it cannot accompany status or parent intent. Malformed integers are syntax
errors; values outside 0–100 are validation errors. Equal patches do not append
history. Moving a subtree checks cycles and depth. Reopening a done task uses
`--status todo` or `--status in-progress`; descendants retain their status and
parent progress is recalculated. Completion may affect ancestors under the
configured parent policy.

## Deletion and terminal output

Interactive deletion requires stdin, stdout and stderr to be terminals. It shows
the target and scope and accepts only `y` or `yes` (case-insensitive); empty
input, other text and EOF decline successfully with no mutation. Cancellation
fails with exit 1. The preview is revalidated inside the mutation; changed
metadata or subtree membership causes conflict and requires a new invocation.
Scripts, redirected streams and `--json` require explicit `--force`. Force does
not imply recursion. Deletion removes associated history; there is no undo.

Human output uses readable status labels, full IDs and cell-aware Unicode
layout. Narrow terminals stack fields and wrap text. Color is disabled for
redirected stdout, nonempty `NO_COLOR`, or `TERM=dumb`. Redirected tables are
plain TSV. Human output escapes terminal controls, bidi controls and line
separators; JSON preserves the original text using JSON escaping. Titles can
be clipped in terminal columns; use JSON for complete notes and other fields.

## Configuration and dates

| Setting | Precedence |
| --- | --- |
| Database | Nonempty `TUSK_DB_PATH`, then absolute `XDG_DATA_HOME` + `/tusk/tusk.db`, then home + `/.local/share/tusk/tusk.db` |
| Parent completion | `--auto-complete-parent[=true/false]`, then `TUSK_AUTO_COMPLETE_PARENT`, then false |
| Timezone | `--timezone IANA_NAME`, then `TUSK_TIMEZONE`, then system local timezone |

Boolean environment values use Go boolean syntax (`true/false`, `1/0`,
`t/f`, with the accepted case variants). A relative explicit database path is
relative to the invocation directory and is treated literally. Invalid paths,
zones, booleans or storage errors never cause a fallback database. Configuration
is validated only for data commands. The production binary embeds timezone data.

Due expressions support `today`, `tomorrow`, `tonight`, three-letter weekdays
(`mon` through `sun`), positive offsets (`+1d`, `+1w`, `+1m`), ISO dates
(`2026-10-01`) and strict timestamps with an offset. Shell quoting does not add
unsupported date grammar: use `+1w`, not `next week`. Relative dates use one
reference instant and the configured location. Month arithmetic clamps to the
destination month's last day. Missing local wall times fail; repeated wall
times choose the earlier instant. `list --due` selects the whole local civil
day. See [service semantics](service.md).

## JSON contract

Success writes one compact JSON value followed by LF to stdout. Errors write
only a safe diagnostic to stderr. Empty collections are `[]`; absent optional
values are explicit `null`. Timestamps are UTC RFC3339 with fractional precision
when needed. No ANSI sequences, terminal queries or interactive prompts are
emitted in JSON mode.

| Result | Fields |
| --- | --- |
| Task (`add`, `edit`, `done`; elements of `list`) | `id`, `title`, `description`, `status` (strings); `priority` (1–4), `progress` (0–100); `parent_id` (string/null); `tags` (string array); `due_date`, `completed_at` (timestamp/null); `created_at`, `updated_at` (timestamps) |
| Tree | Array of nodes: `task` (complete task), `children` (node array), `depth` (integer, selected root starts at 1). Stored `parent_id` is preserved. |
| Delete | `id`, sorted `deleted_ids`, `deleted_count`, `deleted` |
| Stats | `total`, `by_status` (all four status keys, including zeroes), `done`, `completion_percent`, `overdue`, `completed_last_7_days` |
| History | Array of `sequence` (signed 64-bit integer), `task_id`, `kind`, sorted `changed_fields`, `occurred_at` |

History kinds are `create`, `metadata`, `status`, `move`, `progress`, `rollup`.
Use an integer-preserving JSON decoder for sequence values. Stats completion
percentage uses integer division. Overdue excludes done tasks. The seven-day
count describes currently done tasks completed in `(now - 168 hours, now]`;
it is not an immutable productivity log.

## Exit codes and recovery

| Code | Meaning |
| --- | --- |
| 0 | Success, including a declined interactive deletion |
| 1 | Validation, not-found, conflict, configuration, storage, cancellation or output failure |
| 2 | Unknown command/flag, wrong argument count, malformed flag syntax or incompatible flag combination |

Storage is closed before output. A failed pipe or cleanup after an acknowledged
mutation reports **change committed**. Do not repeat the mutation just because
stdout was lost. Read `tusk list --all --json`, locate the complete ID and use
`tusk history ID --json`. Add the full `tree --json` snapshot when hierarchy
matters. Commands do not automatically retry mutations.

An **outcome unknown** diagnostic means commit may or may not have happened.
Start a new process against the same database and inspect tasks/history before
deciding whether to retry. Preserve database/WAL/SHM files. A deleted task has no
history, so use absence from a full fresh snapshot as deletion readback. Killing
a process differs from graceful cancellation; local tests cover both.

## Verification and handoff

Run `make validate build check-generated` in an owned Kitty window and inspect
the current CLI with an isolated temporary database, as required by `AGENTS.md`.
`GOTOOLCHAIN=go1.25.0 make test build-cli` checks the minimum compiler and five
CGO-free executable/test builds. Cross-builds do not prove native console behavior.

`make bench-cli` builds outside timing and runs the full matrix three times.
Each case retains five warmups and 100 consecutive fresh-process samples. Query
p90/p95/p99/max must be below 15/20/30/50 ms; help/version below 5/7.5/10/15 ms.
Every case in every run must pass. Reports include every sample, target-miss
counts, output sizes, binary hash and host manifest; no outlier is discarded. See
[execution evidence](verification-evidence/004/README.md) for the current result.
`make bench-cli-conditions` separately observes first-use, 10,000 tasks, 1 MiB
notes, a held writer and a throttled pipe. `make profile-cli` is an in-process
diagnostic; it cannot replace process acceptance.

Feature 005 owns TUI registration, consent, refresh and draft handling using the
same service/configuration/outcome contracts. Feature 006 owns completion/man
pages, native Windows/macOS and architecture runtime, hosted checks and release
publication. These remain pending until their own candidate evidence exists.
