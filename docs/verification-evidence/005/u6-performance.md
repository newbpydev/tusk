# U6 performance checkpoint

U6 is locally accepted. The approved TUI, automated correctness checks and final
CLI distribution gate pass. Earlier failed runs below retain their original
results; [u6-acceptance.json](u6-acceptance.json) binds the final evidence.

| Evidence | Result |
| --- | --- |
| Approved calendar TUI matrix | 78/78 case-runs pass; preparation worst p95 12.334 ms, max 12.968 ms |
| Root View | Zero allocations in all 27 case-runs |
| Populated PTY frame | 300 samples + 15 warmups; median 26.154 ms; max 47.737 ms |
| Current CLI, unchanged formatter candidate after owner closed Zed | 84/84; query worst p90 13.873 ms; help/version p90 4.592 ms; all tail guards pass |
| Same-host Feature 004 control | 82/84; two 1,000-task tree JSON p90 misses |
| Full validation | make validate build check-generated passed |
| Minimum Go | Final Go 1.25 full tests/five-target builds passed |
| Review | Completed; one useful test addition verified; actual external serving identity unverified |

The control was an isolated git archive of 4f4873c78c484da12f766e700bb69b92f46335a9,
built and measured through its Makefile with GOFLAGS=-buildvcs=false. No production
data was used. All reports retain the complete three-run matrix, samples and host
metadata. The earlier eager-renderer, no-syntax, text-only, and package-update
reports remain in u6-logs; none was relabeled as accepted.

Historical diagnosis: the package-update run observed three pacman processes and perl consuming CPU.
After the update, a five-second host sample still measured Zed at about 67% of
one core. The owner was asked whether a brief idle-host window is available.
The next reference measurement needs confirmed stable conditions; do not repeat
unchanged runs until one happens to pass, choose a best run, change power profile,
or relax the limits. No other owned test/build/review process may run alongside it.
Quit only the owned TUI first, then use the unchanged make bench-cli target.

The diagnostic profile attributes roughly half of repeated in-process CPU to
service task reads, about 19% to storage open and about 8% to JSON formatting.
These are overlapping call-tree observations, not a fresh-process timing budget.
GC/processor-count diagnostic experiments did not establish a robust fix; runtime
defaults and database safety contracts remain unchanged. Do not claim that host
load explains every miss or that the TUI dependency overhead is zero.

The real app is left open in the owned Kitty window with an isolated database.
Visual approval and captures are in u6-kitty/README.md. Source hashes and exact
remaining obligations are in u6-checkpoint.json. Feature V106, final V110
acceptance and the U6 commit remain open. Native/hosted release proof stays with
Feature 006. No push, PR or release was performed.

Source whitespace checks pass. Verbatim terminal captures/logs and hashed upstream
Glamour golden/Markdown fixtures retain their original spaces and fixture syntax;
they are excluded from the generic whitespace-only diff check. Canonical source
formatting and vendor integrity tests pass.

## Calendar candidate — 2026-09-30

The owner-requested calendar and due-date examples change the candidate binary
to `3f74fd34016e6aaf929ca9f764d975e63a6c87a750cf74aa0f62cb7cfdce3914`. Canonical validation, minimum-Go checks,
real Kitty color/plain interaction checks and the calendar code review pass.
All 78 TUI measurement case-runs pass, including nine calendar navigation cases;
worst synchronous p95 12.334 ms and maximum 12.968 ms. Child-PTY startup median
is about 26 ms (no SLA). The previous 69-case reports remain under `u6-logs/`.
See [calendar receipt](u6-calendar.md) for screenshots and exact verification.

The CLI report is still the pre-calendar 79/84 candidate. No CLI remeasurement
or gate waiver is implied by UI approval. U6/V106/ISS-024 remain open pending
a stable-host window or a justified CLI latency correction. The new calendar
is open in owned Kitty; its final visual feedback is pending.

Owner confirmation, 2026-09-30: “Okay, I approve, it looks good.” Calendar visual
acceptance is complete. The instruction is to resume U6 and finish Feature 005.

The owner clarified that Zed is the terminal host for this session. Its ambient
CPU use is recorded, not removed by closing the session. The product policy
requires no concurrent verification workload. That narrower, explicit wording
now replaces Feature 005's ambiguous “no concurrent workload.” Measure the final
calendar candidate after owned checks stop and the owned TUI quits; retain all
samples and the same thresholds. This supersedes the earlier request for an idle
editor window, which would interrupt the user's working environment.

## Final formatter checkpoint — 2026-09-30

The subsequent calendar matrix passed 82/84. A measured JSON formatter
correction reduced isolated encoding median by 18.8%, with exact wire parity.
The corrected binary's full matrix passed 76/84; all reports remain retained.
See [formatter evidence](u6-json-formatter.md) for the eight failures and exact
binary hash. Canonical validation and full minimum-Go tests/builds pass.
No performance handoff or waiver is assumed: the owner is being asked whether
to keep V106 in Feature 005 or carry it explicitly as a Feature 006 release
blocker. The approved app is open in owned Kitty, with unchanged TUI source.

## Final acceptance in the Konsole session — 2026-09-30

The owner closed Zed and requested another measurement under that changed host
condition. A five-second host sample confirms Zed is absent; ambient desktop
work remains recorded in `u6-logs/cli-konsole-host.json`. No concurrent
verification job ran. Power profile, compiler, binary, fixtures and limits are
unchanged. `make bench-cli` passes all 84 case-runs, retaining 8,400 measured
samples and 420 warmups. The 36 individual p90-target misses remain visible;
every case passes its p90 and p95/p99/maximum guards.

| Workload | Worst p90 | Worst p95 | Worst p99 | Worst maximum |
| --- | ---: | ---: | ---: | ---: |
| Queries | 13.873 ms | 14.533 ms | 19.546 ms | 27.451 ms |
| Help/version | 4.592 ms | 4.866 ms | 6.279 ms | 6.297 ms |

The previous 76/84 matrix is retained in
`u6-logs/cli-latency-before-konsole.json`. This result closes the existing gate;
no threshold waiver or deferral was needed. The same source still matches the
last canonical, minimum-Go and completed review receipts. All 78 TUI case-runs
remain valid for unchanged production TUI source. The approved app is reopened
in owned Kitty using its isolated database. Feature 006 native/hosted evidence
remains separate from this local acceptance.
