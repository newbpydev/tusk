# Tusk TUI

Run `tusk tui` (or `./bin/tusk tui` after `make build`) in a terminal of at least
80 columns and 24 rows. The task list takes the left side; details, notes and
activity take the right. Titles and metadata use separate rows. Tasks appear in
Today, Upcoming, Backlog and Completed groups; overdue tasks belong to Today.
Focus has a visible `>` marker. Selection and buttons remain usable without color.

The TUI needs terminal input and output. Redirected streams and `TERM=dumb` are
rejected before opening storage. Use `tusk list --all --json` for pipes. A nonempty
`NO_COLOR` disables color; an empty/unset `TERM` uses plain presentation when
input and output are terminals. Resizing below 80×24 preserves selection and
drafts, suspends hidden form input, and shows a resize message. Ctrl+C still exits.

## Navigation

| Key | Action |
| --- | --- |
| ↑ / ↓ or j / k | Move through tasks; scroll when details have focus |
| Tab / Shift+Tab | Switch list and details |
| ← / → or h / l | Collapse / expand the selected branch |
| Home / End or g / G | First / last task or detail line |
| PgUp / PgDn | Move by a page |
| 1 / 2 / 3 | All tasks / Today (open tasks due today or overdue) / Done |
| / | Search titles and notes; Enter accepts, Esc restores the earlier search |
| f | Open filters; Tab chooses fields, arrows/Space change choices, Ctrl+S applies |
| Esc while browsing | Clear active filters |
| r | Refresh tasks and selected history |
| a / e | Create / edit a task |
| Space / x | Complete the task and its subtree; reopen a completed task as todo |
| d | Preview deletion |
| ? | Open scrollable help; Esc, ? or q closes help |
| q | Quit browsing |
| Ctrl+C | Cancel work, drain storage operations and restore the terminal |

Search waits 150 ms after typing. Matching descendants keep their ancestor path
visible; an ancestor marked `[context]` is shown for navigation. A matching parent
does not automatically include unrelated descendants. Filters combine status and
priority alternatives with all requested tags and the selected local due day.
The search field and view tabs stay visible when the list scrolls. Switching a
tab preserves search and replaces other filters. Today follows the configured
local date, including midnight changes. Applying custom filters replaces the
selected tab; a due-day filter stays on its resolved date until applied again.
The app refreshes every two seconds; an explicit `r` also refreshes. Failed reads
keep the last snapshot labeled stale and pause writes until a successful read.

Notes support Markdown with inert links and images: the app does not open URLs or
fetch remote content. Code blocks use a consistent text color without loading
language-specific syntax themes. Large or unsupported notes fall back to safe, scrollable
plain text. Rendering happens asynchronously. History is loaded separately and
shows stored event order, changed field names and local timestamps; it is not an
atomic snapshot with the task list.

## Forms and saves

Create offers title, notes, priority, due, tags and parent. Edit also offers status
and progress. Tab/Shift+Tab cycles fields and buttons. Enter advances a single-line
field and inserts a newline in notes. Ctrl+S saves. Keys such as `q`, `d`, `?` and
Space are literal input while editing text. The current field number and due-date
timezone are shown.

- With Due date focused, type a day such as `2026-10-15` or `tomorrow`, or press
  Ctrl+P for the calendar. Use ←/→ for days, ↑/↓ for weeks, PgUp/PgDn for months,
  and `t` for today. Enter chooses the highlighted day; Esc returns to the form
  without changing its text. Choosing a day sets the due time to the end of that
  day in the configured timezone when you save. The task is saved with Ctrl+S.
- With Parent focused, Ctrl+P opens a searchable picker. Choose a matching task
  with the arrows; Root removes a parent.
- Ctrl+U clears the focused due, tags or parent field.
- Open leaf progress accepts 0–99. Parent/done progress is derived and read-only.
  Change parent/status separately from a manually edited progress value.
- Esc closes a clean form. A dirty form defaults to Keep editing; choose Discard
  explicitly to abandon it. Focus returns to the initiating pane.
- Saving admits one mutation. Repeated save keys cannot submit duplicates.
  After a successful save, writes wait for a fresh task read. `Saved; refresh
  failed` means the mutation committed: refresh the data without submitting again.

Unchanged fields are omitted from updates, including raw notes and due expressions.
Stored text that cannot round-trip through the editor is displayed read-only,
including control/tab-containing notes, fields over 64 KiB and notes beyond the
editor's 10,000-line capacity. The original bytes remain intact. Ctrl+E offers an
explicit replacement starting from blank; Cancel is the default. Inspect the
complete note in details or CLI JSON before replacing it. Unsupported or oversized
pastes are rejected as a whole. Pasting through the terminal works; the app does
not invoke an OS clipboard program.

External CLI edits can conflict with an open draft. The app retains the draft and
its original task version and stops submission. Ctrl+R offers explicit reload;
choosing it discards the draft only after a successful read. A disappeared task
is never silently recreated. Review the new state and apply any intended change
as a new action.

## Deletion and uncertain outcomes

Deletion shows the title, full ID, task count and history loss. Cancel is selected
initially. A parent requires a separate unchecked subtree checkbox before Delete
can be selected. Tab chooses controls; Space changes the focused checkbox; Enter
confirms the focused button. A second `d` is not consent. Deleted tasks and their
history cannot be undone. If another command changes the task or subtree after
preview, the app reloads the preview and resets consent to Cancel/unchecked.

If a transaction outcome is unknown, writes stop and the attempted action/draft
becomes read-only. The app never retries that mutation. Reload closes the old
storage connection before reading through a new one at the same database path.
Review what is currently saved, then explicitly acknowledge/discard uncertain
intent before any new action. Readback does not prove who made a change. An
uncertain create may have no confirmed ID; matching titles do not identify it.

If storage cannot close, in-session recovery is blocked. Quit, then inspect through
the CLI, for example:

```sh
tusk list --all --json
tusk tree --json
tusk history TASK_ID --json
```

Use the full observed ID in place of `TASK_ID`. Do not resubmit merely because an
error was shown. Ctrl+C or a process signal can arrive after a commit. After
terminal restoration, diagnostics distinguish earlier acknowledged changes from
an uncertain outcome. A session with an uncertain outcome exits nonzero even if
readback was later acknowledged. Cooperative cleanup waits for admitted work;
a second OS signal can terminate stalled cleanup and requires fresh readback.

## Configuration

The [CLI configuration rules](cli.md#configuration-and-dates) also govern the TUI:
`TUSK_DB_PATH`, `XDG_DATA_HOME`, `TUSK_TIMEZONE` / `--timezone`, and
`TUSK_AUTO_COMPLETE_PARENT` / `--auto-complete-parent`. Invalid configuration never
falls back to a different database. Relative database paths resolve once when
opened, so recovery cannot switch paths. Dates accept `today`, `tomorrow`, weekdays,
positive offsets such as `+1w`, ISO dates and timestamps with offsets.

For an isolated demo, set `TUSK_DB_PATH` to a database inside a fresh temporary
directory before creating tasks or launching the TUI. Automated tests/benchmarks,
child-PTY checks and real Kitty observations are separate evidence. Linux terminal
acceptance is local; native macOS/Windows release proof remains Feature 006.
