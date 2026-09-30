# U6 owned Kitty verification

After the final passing CLI matrix on 2026-09-30, the same approved binary was
reopened in owned Kitty window 2 against `u6-tasks.db` at 120x40. The workspace
and saved calendar task remain visible. No test, build or measurement output
was displayed in Kitty; the existing visual captures below bind the same binary.

## Final formatter readback — 2026-09-30

`json-formatter-cli.png` and its text capture show the actual updated CLI in
owned Kitty reading the calendar task from the isolated database. Its identity
and exact stored due timestamp are unchanged. `json-formatter-app.png` shows
the approved Spacious workspace reopened after measurement. No automated test
or benchmark output was displayed in Kitty. The current binary is
`654c9a0d2275d44e3e4da48f97009239da6956e0fe074e7df944de0e5c952fd0`.

## Earlier workflow acceptance

Operator: Codex, 2026-09-29 (America/Sao_Paulo). The owner explicitly approved
the Spacious direction: “Yes, keep this direction.” Preserve padding, two-row
tasks, quick tabs, checkboxes and opaque dialog backgrounds.

Only the real CLI/TUI and a deliberately fault-injected interactive app were
operated in Kitty. Canonical builds, tests and timing ran separately in Bash.
The owned window used an isolated temporary database; user terminals were not
reused or closed.

## Observations

- Spacious workspace, search, All/Today/Done tabs and details at 120×40.
- 80×24 forms, filters, help paging, explicit recursive-delete checkbox and
  buttons remain reachable and legible. NO_COLOR preserves focus/selection.
- Create/paste Unicode text, edit/save, complete/reopen, parent moves, external
  CLI conflict, delete and fresh readback were exercised in this U6 pass.
- 79×23 blocks application edits; returning to supported size preserves drafts.
- Exact 200×60 workspace was checked after correcting the earlier 59-row capture.
- Final help paging, clean quit, visible CLI usage and app relaunch were repeated
  after the Markdown startup correction. before.stty and after.stty are identical.
- Unknown-create injection was operated at 120×40 and 80×24: reload, fresh-owner
  readback and explicit acknowledgement reveal exactly one committed task.
  This is fault-fixture evidence, separate from the production application.
- No reported offset shadow/background artifacts remain in the captured panels
  or dialogs. Screenshot files record the actual Kitty window.

## Candidate identity

Final production recheck binary SHA-256: `c0237537f54b205da76b7aabb3a9b56813019bbcf76a6170ad5fbcacd569b860`.
Earlier workflow captures preceded the startup-only renderer changes. Those
changes preserve the approved layout; final captures are workspace-final-120,
workspace-200, help-80, cli-restored and recovery-120/recovery-80. The recovery
captures use the separately built interactive fixture.

The benchmark phase quits only this owned app to avoid its periodic refresh
competing with measurements. Native macOS/Windows and hosted release checks
remain Feature 006 obligations.

Final owned-window workspace/create-dialog inspection was repeated after review;
see approved-workspace-120.png and approved-dialog-120.png. The empty form was
canceled without saving, and the real production app was left open for the owner.

### Calendar refinement — 2026-09-30

The [calendar receipt](../u6-calendar.md) records 80×24/120×40 color and plain
inspection, six-row month geometry, undersize preservation, cancel/choose/clear,
local end-of-day persistence and fresh CLI readback. Captures/transcripts use the
`calendar-` prefix. The current calendar is open for owner confirmation.

Owner confirmation, 2026-09-30: “Okay, I approve, it looks good.” Calendar visual
acceptance is complete. The instruction is to resume U6 and finish Feature 005.
