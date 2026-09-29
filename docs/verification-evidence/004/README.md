# Feature 004 execution evidence

## Final local branch review

[Review](review-final.md) and [structured receipt](review-final.json): two P2
runtime/verification documentation findings fixed; final diff review clean.
Fresh canonical gate passes, with all 158 accepted source hashes preserved.
The review records independent-peer availability and local coverage explicitly.

## Branch simplification follow-up

[Review and experiment](simplify-review.md): one tree-buffer trial reduced
allocation by about 50% but missed one latency case while the pre-change control
passed. The trial was reverted; accepted production/test source is unchanged.

## Local acceptance: U6 complete

[Acceptance receipt](u6-acceptance.json) · [84/84 latency case-runs](u6-latency-accepted.json)
· [Code review](u6-review-acceptance.json) · [Peer dispositions](u6-review-dispositions.json)

Automated verification ran in Codex Bash. `make validate build check-generated`
and Go 1.25 full tests/five-target cross-builds passed. Query p90 is below 15 ms
in every case/run (worst 14.621 ms); help/version p90 is below 5 ms (worst 2.855 ms).
All tail guards pass. All 8,400 samples, including 40 query target misses, remain
in the report. See the synchronized triplet for the owner-delegated policy.

[Fresh Kitty screenshot](u6-acceptance-kitty.png) and
[transcript](u6-acceptance-kitty.txt) confirm tree/progress/stats behavior.
Local V01–V89 and U6 are accepted. V90–V91 remain Feature 006 native/hosted gates.
Historical failed reports below keep their original outcomes and policies.


## Historical U6 checkpoint: Bash and terminal evidence

[Receipt and exact source hashes](u6-bash-checkpoint.json). Canonical validation,
generated checks, Go 1.25 short tests and five target builds pass in Codex Bash.
[Visible Kitty evidence](u6-bash-checkpoint-kitty.png) and its
[transcript](u6-bash-checkpoint-kitty.txt) verify tree/progress/statistics behavior.
AGENTS.md now reserves Kitty for required visible terminal scenarios.

[Balanced Bash latency](u6-latency-bash.json) and
[temporary performance-profile latency](u6-latency-bash-performance.json) still
fail the unchanged strict gate. U6 is incomplete and uncommitted. The prior
code-review receipt predates the broader changes; fresh review remains required.

PGO, JSON row transport, reader connection reuse and enlarged kernel-pipe
experiments were discarded. Their retained reports/logs are diagnostic evidence,
not acceptance or production build settings. The
[simplification pass](u6-simplify-broad.md) was sequential and inline.

## Historical broader-optimization checkpoint — latency still pending

The owner retained 15 ms. Query allocation, sqlc row loading and stable sorting
have been optimized, and the pipe consumer now prepares storage and collects
prior validation garbage before timing. Every measured sample is still retained.

- [Current receipt and source hashes](u6-broader-checkpoint.json).
- [Canonical gate](u6-checkpoint-gate.log.gz): validate/build/generated checks pass.
- [Minimum Go](u6-checkpoint-minimum.log.gz): full tests and five cross-builds pass.
- [Post-sort balanced run](u6-latency-index.json): 1,000-task query medians about
  9–11 ms; JSON list/tree maxima 17.752/18.818 ms. The strict gate still fails.
- [Go 1.25](u6-latency-index-go125.json), [affinity experiment](u6-latency-affinity.json),
  and [temporary performance profile](u6-latency-performance.json) also fail.
  The last JSON-list maximum is 26.164 ms; the power profile was restored to balanced.
- [Timing diagnostics](u6-timing-scheduling.json), [GC trace](u6-timing-trace.json),
  and [instrumented stages](u6-stage-diagnostic.json) are supplementary evidence.
  Their settings/fixtures differ from reference acceptance; no outcome overrides it.
- [Fresh Kitty screenshot](u6-current-kitty.png) and [terminal text](u6-current-kitty.txt)
  confirm decimal parsing, syntax exit and persisted readback with a temporary DB.

U6/Phase 4 remain unaccepted and uncommitted. The completed review below predates
these optimizations; review the current delta before committing. No new owner
decision about relaxing the bound is needed: keep 15 ms and continue the active
U6 investigation. Do not start Phase 5. The older checkpoint below is historical.

## Earlier U6 checkpoint — before broader optimization

Kitty and Git access are working. The mandatory Kitty rule is in `AGENTS.md`.
Six units are committed in the required order; U6 functional verification passes,
but **V87 / ISS-023 fails**. U6 and Phase 4 remain unaccepted; Phase 5 must not start.

- [U6 receipt and source hashes](u6-checkpoint.json), [CLI guide](../../cli.md).
- [Canonical gate log](u6-review-green.log.gz): `make validate build check-generated`
  passes, including races, 29 script fixtures and coverage (CLI 96.4%, main 95.7%,
  storage 97.6%, benchmark runner 95.3%).
- [Minimum Go final log](u6-minimum-final.log.gz): full tests and five CGO-free
  executable/test builds pass. These builds do not establish native target runtime.
- [Final benchmark](u6-latency-final.json): all 100 samples per case retained.
  Help/version, empty and 100-task cases pass. 1,000-task JSON list/tree maxima are
  **30.118/32.832 ms**, each with 100/100 violations of the unchanged 15 ms bound.
  Human list/tree and some stats samples also fail. Earlier baseline, minimum-Go
  and optimized reports remain alongside it. A failed gate cannot be accepted by
  choosing a different percentile or dropping outliers.
- [Separate capacity/conditions observations](u6-conditions-final.json): first-use,
  10k tasks, 1 MiB notes, held writer and throttled output, each with a deadline.
- [Completed code review](u6-review.json) and [resolution](u6-review-resolution.json):
  decimal progress and incomplete benchmark fixtures fixed after observed red;
  performance finding remains open. Local review ran sequentially in the main
  agent. The attempted peer model had no verifiable independence receipt.
- Actual Kitty screenshots: [40 columns/color](u6-width40-color.png),
  [40 columns/dumb](u6-width40-dumb.png), [80 columns/no color](u6-width80-no-color.png),
  [120 columns/color](u6-width120-color.png), [final decimal/readback](u6-final-kitty.png).
  Matching text captures are retained. These are distinct from automated real-PTY
  yes/no/EOF/SIGINT/SIGTERM tests and process broken-pipe/hard-kill readback.

Timestamp decoding and removal of a redundant output copy reduce allocation;
regression and canonical checks pass. The final process benchmark still fails.
Further measured optimization or an explicit owner decision is needed. No revised
bound, acceptance, native/hosted proof or publication is inferred.

## Prior unit evidence

## U1 locally verified — 2026-09-28

The session permission issue is resolved. [U1 acceptance](u1-accepted.json)
records passing canonical gates in Kitty and retained minimum-Go/five-target
build evidence. [Actual terminal screenshot](u1-kitty.png) and
[screen text](u1-kitty.txt) show help, version and syntax-error exit 2. The earlier
blocked checkpoint below is historical and does not describe current permissions.
U1 advances to its local commit before U5 begins.

## Historical U1 checkpoint — 2026-09-28

U1 is implemented locally but **not accepted or committed**. U5 has not started.
The [receipt](u1-checkpoint.json) identifies the baseline, source file hashes,
commands, results and blockers. Since the source is uncommitted, the baseline SHA
alone does not identify the tested implementation; use the source manifest.

Observed red evidence is retained for unknown-command exit 0 (expected 2), missing
CLI/composition APIs, and the file-only Make build omitting a sibling source.
The copied build fixture later exposed a sandbox VCS-discovery failure; disabling
VCS stamping for that disposable non-repository fixture resolved it. Production
builds still use the normal Go VCS behavior.

Passing **headless** checks:

- `make validate build`: formatting, vet, full tests, race tests, coverage and
  script fixtures; CLI coverage 98.4%, composition/main coverage 100%.
- Go 1.25 full tests, then final CLI tests and complete executable/test builds for
  Linux amd64/arm64, macOS amd64/arm64 and Windows amd64 with CGO disabled.
- Dependency graph preserves SQLite v1.58.0 and libc v1.75.6. Cross-build logs
  retain nonfatal module stat-cache write warnings from the restricted session.

Required **Kitty verification is pending**. Both launch logs report a display
connection failure before the verification script ran: Wayland connection failed;
X11 could not open `:0`. No visual acceptance or native non-Linux acceptance is
claimed. The user-required Kitty rule is now in `AGENTS.md`.

The follow-up [permission diagnosis](session-permissions.json) establishes the
cause: Kitty 0.49.1 is installed, both display sockets exist and Xauthority is
readable, but both socket connections return `EPERM` (errno 1). The session
sandbox denies desktop access. `.git` is also mounted read-only. Installing Kitty
or changing display settings does not address these permission boundaries;
the session must permit desktop connections and Git metadata writes.

The implementation was staged. A subsequent attempt to unstage only the agent's
paths failed because `.git/index.lock` could not be created on the read-only
filesystem. Later planning/evidence changes are unstaged. No commit or publication
occurred. Preserve this index and working tree when resuming.

Resume in a session with desktop and normal Git access. Run
[kitty-verify.sh](kitty-verify.sh) in an owned Kitty window, inspect help, version
and syntax-error output, retain the visible terminal result, complete the U1
review, synchronize the triplet/masterplan, then make the U1 commit before U5.
The script uses the temporary Go cache created during this run; use another
writable Go cache if resuming after temporary files were removed.

## U5 locally verified

Explicit JSON DTOs preserve complete task and query data, nulls and arrays, UTC times and exact history sequence numbers. Output failures retain known committed state; unknown outcomes and private error redaction remain intact.

See [receipt](u5-accepted.json), [Kitty screenshot](u5-kitty.png) and
[screen text](u5-kitty.txt). Canonical gates passed; commit precedes U4.

## U4 locally verified

Human output escapes terminal controls and bidi directives, preserves complete IDs, wraps by grapheme cell width, and uses invocation-owned styling without background probes. Actual Kitty inspection found and fixed header alignment; the six-cell terminal priority header is PRIO, while TSV remains PRIORITY.

See [receipt](u4-accepted.json), [Kitty screenshot](u4-kitty.png) and
[screen text](u4-kitty.txt). Canonical gates passed; commit precedes U2.

## U2 locally verified

Add, edit and done map exact argument intent to one accepted service call. Disk integration proves subtree lifecycle, atomic combined patches, no-op history, parent policy, and supported service dates (+1d/+1w/+1m). The built CLI was exercised in Kitty on an isolated database. Shared-code reuse, quality and efficiency review ran sequentially with no behavior-preserving change warranted.

See [receipt](u2-accepted.json), [Kitty screenshot](u2-kitty.png) and
[screen text](u2-kitty.txt). Canonical gates passed; commit precedes U7.

## U7 locally verified

Deletion requires independent recursion and force intent, defaults to no in an eligible terminal, and passes unchanged preview consent to the service. Second-owner add/remove/move/metadata races reject stale consent. Unix input uses bounded polling; Windows cancellation joins its pinned reader, including the no-pending-I/O race. Actual Kitty decline, acceptance and recursion refusal passed.

See [receipt](u7-accepted.json), [Kitty screenshot](u7-kitty.png) and
[screen text](u7-kitty.txt). Canonical gates passed; commit precedes U3.

## U3 locally verified

List, tree, stats and metadata history consume authoritative service read models without resorting or recomputing. Tests cover exact filters, empty/missing results, retained completions and failure suppression. Actual Kitty query inspection improved long history output to labeled records when aligned columns cannot fit.

See [receipt](u3-accepted.json), [Kitty screenshot](u3-kitty.png) and
[screen text](u3-kitty.txt). Canonical gates passed; commit precedes U6.
