# PR #5 review remediation

### PR #5 hosted review R1 (2026-10-01)

Three stale-write UX findings are fixed with observed failing regressions:
paused form saves show an in-modal Ctrl+R hint; a paused delete supports r and
Ctrl+R, revokes old consent and loads a fresh preview; rejected toggles pause
writes until successful readback. Canonical `make validate build` passes.
Owned Kitty inspection at 100x32 confirms visible stale-form feedback and
retained raw draft. Delete-read abandonment and renewed consent are automated
regressions, not additional Kitty evidence. Receipts: `docs/verification-evidence/005/review-pr5/`.
Remaining hosted threads and Feature 006 native/release gates stay open.

### PR #5 hosted review R2 (2026-10-01)

Seven navigation/input/presentation findings have regression coverage and fixes:
paging counts task rows and group gaps with visible overlap; accepted search is
trimmed; collapse/projection share the end-inclusive predicate; search and filter
paste is atomically bounded and control-checked; picker windows start on titles
and retain the selected ID; details no longer render the form-template editor;
zero due dates display Not set in summary and recovery. Picker offset rounds up
to the next title, rather than down, so both selected lines remain visible at odd
heights. `make validate build` passes. Owned Kitty at 100x32 confirms whitespace
search acceptance, filter presentation and clean quit against an isolated DB.
The focused regression and Kitty receipts are under `review-pr5/`. Lifecycle and
portability review remains open; native/hosted release gates remain Feature 006.

### PR #5 hosted review R3 (2026-10-01)

Thirteen lifecycle, environment, dependency and verification findings are fixed.
Run rejects missing Open before terminal startup; the cold model waits for actual
terminal dimensions; SIGHUP follows graceful close and terminal restoration.
Display escaping covers invisible format controls while preserving ZWJ emoji.
The CLI and TUI use the same injected environment. Exact replacement pins now
carry maintenance guidance. Reachable dependency advisories are removed with
x/net 0.55.0, x/text 0.39.0, goldmark 1.7.17 and x/term 0.43.0.

Tests locate the source checkout independently of cwd, use the physical temporary
path and HOME/USERPROFILE, embed test-only tzdata, allow a five-second stress
watchdog and benchmark escaped JSON separately. The manual Kitty fixture is a
standalone, explicitly tagged binary, with no testing-runner os.Exit or weakened
coverage gate. The R2 screenshot was replaced with the inspected filter capture.

Observed Red/Green, final `make validate build check-generated bench-cli-json`,
minimum-Go focused race tests and five-target cross-builds pass. Owned Kitty
confirms the small-terminal banner, restored shell after SIGHUP and standalone
fixture quit. The current TUI matrix passes 78/78 cases plus startup. Its runtime
source matches; the subsequent Makefile-only change selects the escaped JSON
microbenchmark and does not affect the TUI binary or measurement.

`GOTOOLCHAIN=go1.25.13 make check-vulnerabilities` reports zero reachable or
imported-package vulnerabilities and one uncalled module advisory. The default
custom Go 1.27 compiler is unsupported by the installed source analyzer; its
failure is retained. Go 1.25.0 compatibility tests are not a security claim for
that old compiler. CLI acceptance remains open: first candidate 82/84,
unchanged earlier-commit control 80/84, quiet candidate 79/84. Every failed case
and raw sample is retained; no historical pass substitutes for current proof.
A patched-compiler comparison is underway. Hosted settlement and native/release
proof remain pending. Receipts: `docs/verification-evidence/005/review-pr5/`.

### PR #5 hosted follow-up R4 (2026-10-01)

Four follow-up findings reproduce and are fixed: successful forest refresh clears
recognized write-pause/read-failure form feedback without clearing validation
errors or drafts; successful preview clears completed retry feedback and revokes
old consent; accepted filter edits/navigation clear rejected-input feedback;
failed-read abandonment uses dialog-neutral before retrying wording.
Focused Red/Green and `make validate build check-generated build-tui-fixture`
pass. Owned Kitty confirms the pause disappears after Ctrl+R, retains the raw
draft and quits cleanly. The screenshot was visually inspected; the initial
Wayland capture could not be taken by the X11 capture tool, so a fresh owned X11
window repeated the app check. No automated results were displayed in Kitty.

The Go 1.25.13 CLI comparison also passes only 79/84; a temporary performance
profile hold for the unchanged control passes 83/84 (tree p90 15.117 ms in run 1).
The hold ended with that command and balanced is restored. A diagnostic-only
nine-mode fresh-process sample set retains wall/child CPU times and parent GC
counts. Ordinary child trace produces no GC line in the inspected sample;
changing child GC/procs does not establish a uniform latency improvement.
No runtime knob, output contract, threshold or failing sample was changed.
Current CLI acceptance and hosted settlement remain open; a final candidate
matrix follows this committed feedback unit. All reports are under review-pr5.

### PR #5 hosted follow-up R5 (2026-10-01)

The additional invisible-text finding is fixed after eight observed regression
failures: display escaping covers Mongolian vowel separator, both jamo fillers,
full/half-width Hangul fillers and the complete U+2060–U+206F format/reserved
range. Stored content is untouched; ordinary Hangul, ZWJ emoji and ZWNJ script
composition retain their existing display behavior. This bounded display policy
uses visible representations of default-ignorable characters, as allowed by the
[Unicode display FAQ](https://www.unicode.org/faq/unsup_char.html); reserved ranges
are described in [Unicode 16 chapter 5](https://www.unicode.org/versions/Unicode16.0.0/core-spec/chapter-5/).

Focused Red/Green and `make validate build check-generated` pass. An isolated
owned Kitty CLI window confirms visible escapes and normal Hangul/emoji; its
capture was inspected and the window closed. The first combined fixture title
hit the existing table title limit, so two short titles provide complete visible
coverage without changing the formatter. Kitty's text dump escapes the joined
emoji's ZWJ; the native screenshot and raw-string unit test cover that separately.

The R4 candidate CLI matrix remains failed at 82/84, with large JSON-tree p90
19.475 ms in run 1 and 16.750 ms in run 2. Every sample is retained. CLI latency
acceptance and hosted settlement remain pending after this committed review fix.

### PR #5 hosted follow-up R6 (2026-10-01)

The browse-footer follow-up reproduces for both delete-refresh keys. Successful
forest readback now clears only the exact failed-read notice, alongside its modal
counterpart. A separate regression preserves unrelated write-receipt notices.
Focused Red/Green and `make validate build check-generated` pass. This message
state change is covered by automated regressions; prior owned Kitty receipts
cover the actual form-refresh flow, without claiming a new delete-fault run.

The performance-profile candidate also fails (80/84); the temporary hold ended
and balanced is restored. Failed cases include three large JSON-tree cases and
one large JSON-list case; a tree case with p90 below 15 ms still fails its p95
limit. All samples/outliers remain retained. No performance acceptance or merge
readiness is claimed. Review fixes are committed independently of the remaining
CLI performance investigation and hosted settlement. The owner has been asked
whether to continue that investigation or finish review handling and pause with
latency blocked. Native/hosted release proof remains Feature 006.

### PR #5 hosted follow-up R7 and receipt correction (2026-10-01)

The reviewer's explicit-exemption alternative is adopted for shaping marks and
variation selectors. Passing-before-change regressions pin all cited ranges and
composed Mongolian, Khmer, combining-mark, ideographic and emoji strings. The
sanitizer comment documents their retention and the use of opaque task IDs for
identity. This is a policy/test clarification with no runtime behavior change,
so no failing runtime defect is claimed. Glyph-selection rationale is grounded
in the [Unicode variation FAQ](https://www.unicode.org/faq/vs.html) and
[UTS #37](https://www.unicode.org/reports/tr37/).

Receipt correction: both R4 fixture quit attempts stopped at the dirty-draft
Discard prompt. Their raw draft/pause-refresh captures remain valid, but the
original clean-quit statement was premature. Both owned fixtures were later
explicitly discarded (Tab, Enter) and quit (q); process exit was verified and is
retained in r4-kitty-quit-correction.json. They ran during the later CLI matrices;
the initial R3 failures predate them. All failed matrices remain unchanged. The
next balanced-profile CLI matrix will run with these owned fixtures closed.

The fresh R6 TUI matrix passes 78/78 cases plus startup, with its full source
manifest and samples retained. R7 changes only a comment and test coverage;
production logic matches that measured R6 runtime. Minimum-Go focused race tests
also pass. `make validate build check-generated` passes for R7 before commit.
CLI acceptance and hosted settlement remain pending; native/release proof stays
with Feature 006.

### PR #5 hosted follow-up R8 (2026-10-01)

Red tests reproduce stale browse-footer notices after conflict, missing-task and
storage-failure toggle readback, plus a confirmed form reload. Notices now carry
a refresh-dependent lifecycle set by their source, instead of matching one error
string. Replacing a notice resets that lifecycle; successful matching readback
clears refresh-dependent failures while committed-write receipts survive. Failed
retries retain their notices. Existing tests cover committed-save refresh failure
and recovery to Saved, alongside unknown outcomes and generation checks.

`make test-unit` Red/Green and `make validate build check-generated` pass.
No new native terminal fault injection is claimed. The complete balanced CLI
matrix with all owned fixtures closed remains 82/84 under the original target;
its slowest case averages 15.041 ms p90 across three runs. All samples remain
unchanged. The owner authorized continued investigation, then explicitly allowed
an average-based host calibration. That separate verification-contract unit follows; current latency acceptance and hosted settlement remain open.

### Owner-authorized PR #5 host calibration (2026-10-01)

The reference CLI protocol remains the default. The owner authorized using the
average speed of the runs to calibrate acceptance for this computer. For Feature
005 PR #5 on Linux amd64, AMD Ryzen 5 4500U, balanced profile, use
`make bench-cli CLI_BENCH_PROFILE=ryzen-4500u-balanced-v1` with an explicit output
path. Each of the 28 command/fixture cases must have exactly runs 1, 2 and 3,
five warmups and 100 valid consecutive samples per run. Its arithmetic mean of
the three p90 values must be below 18 ms for queries and 5 ms for help/version.
Every run independently meets host p95/p99/max guards: queries below
22/30/50 ms; help/version below 7.5/10/15 ms. Errors, incomplete matrices,
duplicate runs and invalid samples fail. This does not average away tail failures.

The 18 ms query ceiling rounds the largest retained candidate three-run mean
(16.787 ms) upward with about 7% margin; the latest fixture-closed tree mean is
15.041 ms. Hardware capacity is a hypothesis, not a proven cause. No child
runtime knobs, workload sizes, timings, outliers or production code change for
this calibration. Original case `passed`, percentiles, 15/5 ms limits and miss
counts remain visible. `host_acceptance` separately records the chosen policy,
case means and result. Old reports remain unchanged and failed. The fresh host
matrix and current TUI readback matrix must be retained before acceptance.

Calibration refinement before acceptance: the initial host matrix meets mean
p90 (worst 16.524 ms) but fails one 20 ms p95 guard at 20.093 ms. It remains
failed in host-cli-latency-initial.json. The retained earlier control p95 reached
21.896 ms; a rounded 22 ms host p95 guard covers that observed range. Only the
host query p95 guard changes; reference limits/results, help/version, query
p99/max, sample counts and all outliers stay intact. A fresh complete matrix is
required against this revised contract. This is explicit owner-authorized host
calibration, not a measured production speedup or proof of hardware causation.

Host calibration closure: the fresh complete CLI matrix passes all 28 host
case groups. Worst query/help three-run mean p90 is 16.189/3.786 ms. Query
p95/p99/max peaks at 19.727/21.095/22.509 ms. The original reference assessment
remains failed at 82/84; both assessments and every sample are retained in
host-cli-latency.json. The initial calibration failure is unchanged. The fresh
TUI matrix passes 78/78 plus three startup runs. All share the same measured
binary hash; host-acceptance.json pins runtime and benchmark sources, including
the R8 notice fix. Only subsequent evidence/documentation changes may reuse this
source proof; commit/VCS metadata changes are not a new runtime measurement.
`make validate build check-generated` passes, with benchmark-package coverage
95.5% after closing the initially missed reporting path (93.8%). This closes
local acceptance under the explicit host exception; hosted settlement is next.
