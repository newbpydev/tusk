# PR #4 review remediation, round 1

Baseline: `fd222a9204af33b595e830cd4e8637a535b7c667`. The user invoked
`ce-babysit-pr 4` on 2026-09-29; fixes, commits, pushes and review replies are
authorized, while merge remains user-owned. Review was evaluated sequentially
under AGENTS.md. No Feature 005 implementation was started.

## Dispositions

All 21 Kilo threads were investigated: 14 fixed, three fixed differently, three
declined to preserve accepted contracts, and one disproved by a subprocess test.
Thread reply/resolution is performed after the validated commit is pushed.

| Thread | Verdict | Evidence and decision |
|---|---|---|
| [4128696952](https://github.com/newbpydev/tusk/pull/4#discussion_r4128696952) | fixed | The human-output sanitizer now escapes U+061C, U+200E, U+200F and U+FEFF. Regression tests assert the exact visible escapes; JSON continues to preserve the original values. |
| [4128696956](https://github.com/newbpydev/tusk/pull/4#discussion_r4128696956) | fixed-differently | The process now restores default signal handling before publishing first-signal cancellation. A second interrupt can terminate stalled console cleanup. A real subprocess test injects a permanently failing cancel callback and verifies both SIGINT and SIGTERM escalation. Normal cleanup still joins the pinned reader: a timeout that returns while stdin is still being read would violate V57 and could consume later input. Native Windows console execution remains the separate V90 release obligation. |
| [4128696962](https://github.com/newbpydev/tusk/pull/4#discussion_r4128696962) | fixed | Moved empty-filter semantics beside TaskFilter as HasPredicates and made the SQLite fast path use it. Tests cover each of the eight predicates and empty slices. This centralizes maintenance without changing query results. |
| [4128696966](https://github.com/newbpydev/tusk/pull/4#discussion_r4128696966) | fixed | Corrected all seven directly imported modules, including x/term and x/sys, and synchronized go.sum. make check-modules now exposes go mod tidy -diff as a canonical check; no dependency version changed. |
| [4128696969](https://github.com/newbpydev/tusk/pull/4#discussion_r4128696969) | fixed | Child GOGC/GOMAXPROCS modes now leave parent GC and pipe allocation at baseline. The explicit combined-parent mode remains available. All nine mode selections have regression coverage. Historical u6-timing artifacts are retained, with documentation that they cannot establish child-only effects. |
| [4128696973](https://github.com/newbpydev/tusk/pull/4#discussion_r4128696973) | fixed | Default handling is restored before the cancellation context becomes observable. Subsequent SIGINT/SIGTERM can terminate stalled cleanup; the cleanup callback joins the signal goroutine. Subprocess tests cover a permanent console-cancel failure. |
| [4128696977](https://github.com/newbpydev/tusk/pull/4#discussion_r4128696977) | declined | Retaining default-no EOF handling. Feature 004 KTD5/V56 and TestConfirm_Lines explicitly require an unterminated yes to decline. Treating EOF as affirmative would expand destructive consent beyond the accepted contract; Enter submits y/yes. |
| [4128696981](https://github.com/newbpydev/tusk/pull/4#discussion_r4128696981) | fixed | Added trusted diagnostics naming TUSK_AUTO_COMPLETE_PARENT or --timezone/TUSK_TIMEZONE with valid-setting examples. Regression tests prove rejected values remain private and invalid configuration opens no service. |
| [4128696986](https://github.com/newbpydev/tusk/pull/4#discussion_r4128696986) | fixed-differently | Kept the three-stream TTY requirement and stderr-only prompt contract. The error now offers a safe interactive recovery: omit --json and restore stdin, stdout and stderr to the terminal. It still documents explicit --force for automation, without silently moving prompts to stdout. |
| [4128696990](https://github.com/newbpydev/tusk/pull/4#discussion_r4128696990) | fixed | The CLI now maps ErrChildrenPresent to a trusted message explaining --recursive and whole-subtree deletion. The service/storage sentinel and the pre-prompt/transaction checks are unchanged. |
| [4128696995](https://github.com/newbpydev/tusk/pull/4#discussion_r4128696995) | declined | Keeping transaction-uncertainty precedence from KTD7/V25. A known sentinel inside an errors.Join tree does not prove that the TransactionError is harmless or pre-mutation; prioritizing it could hide a genuinely uncertain commit. Production Close does not return transaction uncertainty, and the existing joined/value/pointer tests pin the conservative contract. |
| [4128696999](https://github.com/newbpydev/tusk/pull/4#discussion_r4128696999) | fixed | Empty/unset TERM, including an absent environment provider, now selects plain output. ANSI requires a TTY, nonempty non-dumb TERM and no nonempty NO_COLOR. The capability matrix covers empty TERM. |
| [4128697002](https://github.com/newbpydev/tusk/pull/4#discussion_r4128697002) | fixed | Clarified TaskReader: each call transfers exclusive ownership of the result slice and every mutable field, even across calls inside one snapshot. Callers may compact, reorder and clear results. Existing real-repository tests cover detached data. |
| [4128697007](https://github.com/newbpydev/tusk/pull/4#discussion_r4128697007) | fixed | Added an explicit reversed-root/reversed-child fixture. It asserts canonical priority/ID ordering and recursive depth independently of CompareTasks, exercising the reorder branch. |
| [4128697013](https://github.com/newbpydev/tusk/pull/4#discussion_r4128697013) | fixed-differently | Documented catalog regeneration in the SQLite upgrade procedure. TestEmbeddedSchemaCatalog already evaluates every migration prefix with the linked runtime and compares all generated objects, so make validate and make check-generated reject normalization drift before release. Keeping that enforced regeneration boundary avoids adding per-open version probing or a second catalog-validity mechanism. |
| [4128697018](https://github.com/newbpydev/tusk/pull/4#discussion_r4128697018) | not-addressing | The panic premise does not hold: os.ProcessState.ExitCode explicitly returns -1 for a nil receiver. The current code then fails through t.Fatalf because -1 differs from the expected exit. A new subprocess regression with a readable non-executable binary confirms a controlled first-use launch failure and no panic, on both tested Go versions. |
| [4128697023](https://github.com/newbpydev/tusk/pull/4#discussion_r4128697023) | fixed | Added snake_case tags to all case and top-level reference report fields, including percentile units and limits. An exact-key test protects the emitted schema. Historical artifacts keep their original keys; docs describe that distinction. |
| [4128697031](https://github.com/newbpydev/tusk/pull/4#discussion_r4128697031) | fixed | Negative measurement-target invocations now pass BUILD_OUTPUT inside the test temporary directory, so their build prerequisite no longer creates the repository bin directory. |
| [4128697037](https://github.com/newbpydev/tusk/pull/4#discussion_r4128697037) | fixed | The compiled profile test binary now uses the selected CLI_PROFILE_OUTPUT plus .test. A dry-run fixture checks the path, including spaces; callers can choose a writable directory for both artifacts. |
| [4128697047](https://github.com/newbpydev/tusk/pull/4#discussion_r4128697047) | fixed | Added a shuffled fixture with a handwritten expected priority/ID order and stable equal-key rows. The randomized parity test remains useful for implementation parity; semantic direction no longer depends only on that oracle. |
| [4128697050](https://github.com/newbpydev/tusk/pull/4#discussion_r4128697050) | declined | Keeping the allocation regression gates. They protect the measured performance fixes accepted in U6; hiding them behind a tag would remove that protection from make validate. The byte/collection bounds already allow headroom, while the zero-allocation fast paths are explicit requirements. The full suite is checked with both the current toolchain and minimum Go 1.25; a future compiler change should be investigated and measured before a bound is changed. |

## Verification

- `make validate build check-generated check-modules`: passed, including full
  tests, race detection, coverage, 35 script fixtures and generated catalog/sqlc
  checks. No dependency version changed.
- `GOTOOLCHAIN=go1.25.0 make test build-cli`: passed, including five CGO-disabled
  CLI and test cross-builds. Native Windows/macOS execution remains deferred.
- Red/green logs cover directional marks, empty TERM, trusted hints, permanently
  stalled cancellation and second-signal exit, missing filter API, JSON report
  keys, diagnostic-mode isolation, profile output and module metadata. Added
  independent sort/tree fixtures pass without production ordering changes.
- The launch-failure reproduction did **not** panic before the change. Its
  initial test failed only on a proposed message expectation; the final test
  checks the existing controlled failure. No launch-path implementation changed.
- Owned Kitty inspection used an isolated database and verified colored output,
  visible directional escapes, private configuration hints, redirected-prompt
  recovery, interactive decline and plain output with empty TERM. The owned
  window was closed after inspection. See [screen](review-pr4-r1-kitty.png) and
  [text](review-pr4-r1-kitty.txt).
- Reference timing: the isolated matrix passed all 84 case-runs and retained
  8,400 samples. Worst help/version p90: 2.751 ms; query p90: 12.495 ms. All
  tail guards passed; 36 individual target misses remain retained.
  [Raw report](review-pr4-r1-latency-isolated.json.gz),
  [receipt and source hashes](review-pr4-r1.json).
- The first 84-case run overlapped
  the end of compatibility cross-builds and failed six cases. It is retained
  as diagnostic evidence, not acceptance or an outlier-discarding retry.

Historical timing diagnostics retain their original observations but do not
isolate child settings from parent GC/preallocation. New reference reports use
snake_case keys; historical reports keep their original schema. Allocation
regression gates remain mandatory and passed on both tested compilers.
