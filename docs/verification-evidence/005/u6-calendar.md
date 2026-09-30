# U6 due-date calendar — 2026-09-30

The owner approved the existing Spacious design and requested a calendar plus
accepted-input examples. This refinement implements that request in production.
It does not close the existing final CLI latency gate or create a U6 unit commit.

## Behavior and evidence

- Due label: `Due date · e.g. 2026-10-15 or tomorrow`. The adjacent hint shows
  Ctrl+P calendar, Ctrl+U clear and the configured timezone.
- Ctrl+P opens the month containing the parsed due value, or local today for an
  empty/invalid value. Left/right move a day, up/down a week, PgUp/PgDn a month
  with day clamping, and `t` selects today. The grid uses civil dates so DST does
  not repeat or skip navigation days. The service remains the date authority.
- Enter copies a `YYYY-MM-DD` day to the draft; Ctrl+S still saves. Esc preserves
  the original text exactly. Calendar input cannot trigger browsing or save.
  Read-only, conflict, saving and undersize gates remain in place.
- The 72×18 dialog shares the approved opaque surfaces, padding, buttons and
  dimmed backdrop. A bracketed selection and today dot also work without color.
  Root View only returns its prepared immutable string.

Red was observed before implementation: every new calendar test failed with
`due field did not open a calendar` in `u6-logs/calendar-red.log`. Green covers
leap/common years, year boundaries, month clamping, local today/UTC timestamps,
spring/fall DST, cancellation and unchanged edit, keyboard trapping, resizing,
deep View purity, plain/color layout and real SQLite save/readback.

`make validate build check-generated` passes (`u6-logs/calendar-validate.log`),
including race and 97.1% TUI coverage. The first full check caught the old label
assertion; it was updated to require both the example and timezone guidance.
That failed log is retained as `calendar-validate-first.log`. Go 1.25 focused
calendar/form tests and five CGO-free target builds pass (`calendar-go125.log`);
final layout tests also pass on Go 1.25 (`calendar-go125-final.log`).

## Real Kitty inspection

Owned Kitty `final.sock`, window 2, isolated `u6-tasks.db`; existing user tasks
and other terminal sessions were preserved. Only real app/CLI usage ran here.
Automated checks and measurements ran in Codex Bash.

- 80×24: inspected a six-week August 2026 grid, including selected August 31.
- 79×23: resize message replaced the app; right/Enter/Esc were suspended.
  Restoring 120×40 preserved August 31. PgDn clamped it to September 30.
- Esc returned to the original typed `2026-08-31`. Ctrl+U cleared it. Typing
  `2026-10-15`, opening the calendar, moving right and choosing returned
  `2026-10-16` to the form before save.
- Saving created exactly one `Calendar usability check`. After normal quit,
  fresh CLI `list --all --search "Calendar usability check" --json` returned
  `2026-10-17T02:59:59.999999999Z`, the end of October 16 in America/Sao_Paulo.
  Terminal prompt/cursor/input restored normally.
- NO_COLOR at 80×24: selected `[30]`, today dot, weekday grid and controls remain
  readable. Choose/reopen/cancel/clear returned to a clean form without saving.
- Returned to color at 120×40 with the calendar open in a new blank form for
  the owner's requested visual confirmation. No additional persisted task.

Viewed captures and terminal transcripts are in `u6-kitty/calendar-*`.

## Review

`ce-simplify-code` reuse, quality and efficiency lenses ran sequentially inline
under AGENTS. Existing action buttons, modal compositor, surface handling and
injected date parser were reused. No further behavior-preserving simplification
was warranted (0 applied, 0 skipped findings); month arithmetic remains local UI
navigation rather than creating a new service API.

`ce-code-review` completed its lite path for this visible, local calendar delta,
with no findings. It checked the explicit refinement requirements, AGENTS,
form/Update/save call paths, tests and measurement instrumentation. No reviewer
agents or external peer were used. See `calendar-review/review.json`. The review
base is an unreferenced temporary commit object of the previously validated
staged tree; the implementation branch and HEAD never moved. It is not a U6
unit commit. Broader branch review remains in `review-r1/`.

## Remaining acceptance

All 78 final TUI case-runs pass, including nine new calendar navigation cases
alongside the original 69. Calendar preparation worst p95 is 12.334 ms and
maximum 12.968 ms; all existing preparation budgets pass. The 300 retained
child-PTY startup samples have a median of about 26 ms and maximum 47.737 ms
(startup is an observation, not an SLA). The pre-calendar reports are retained under `u6-logs/`.
The last CLI report passed 79/84; its pre-calendar binary identity is retained.
No claim of current-candidate CLI acceptance is made. V106/005-ISS-024 and the
U6 commit remain open pending stable-host evidence or a justified latency fix.

Owner confirmation, 2026-09-30: “Okay, I approve, it looks good.” Calendar visual
acceptance is complete. The instruction is to resume U6 and finish Feature 005.
