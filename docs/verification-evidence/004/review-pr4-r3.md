# PR #4 performance remediation and third review round

Baseline: `b450752ff33f020ea05e62a361a1b343c50f978d`. The owner explicitly
selected continued performance remediation after the round 2 latency misses.
All automated measurements use default Codex Bash. No Kitty window, concurrent
build/test workload, runtime scheduling override, threshold change or sample
trimming was used for the accepted matrix.

## Diagnosis and bounded change

An isolated checkout changed the existing profile benchmark from list JSON to
1,000-task tree JSON and added memory profiling through `make profile-cli`.
The [diagnostic patch](review-pr4-r3-profile-worktree.patch.gz) records that
setup plus the candidate diff; the temporary checkout was removed afterward.
The CPU profile attributes about 11% of samples to background GC. JSON buffer
reservation is the largest single allocation (29% of sampled allocated bytes).
This identifies avoidable allocation work; it does not prove that GC caused
every historical tail miss or exclude host variability.

The encoder reserved maximum fixed widths even when nullable fields were absent
or UTC timestamps had no fractional seconds. The new red test allocated 557,074
bytes for 434,002 wire bytes. Capacity now reflects the fixed wire shape, present
optional values, integer widths and timestamp precision. It remains an estimate:
escaped strings can grow normally. The trailing newline has reserved space.
No task ownership, query, tree construction, JSON bytes or error contract changes.

The new allocation regression passes with a 110% output-size budget. Existing
JSON parity checks cover control bytes, Unicode, escaping, nullable fields,
non-UTC timestamps, deep trees, invalid dates and explicit schema fields.

The diagnostic profile changes from 2,142,906 to 2,034,697 bytes/op (108,209 bytes
less) and 26,019 to 26,018 allocations/op. Its time changes from 9.202 to 9.545
ms/op, so it is **not** claimed as a timing improvement. Fresh-process acceptance
below is the timing gate; these in-process profiles are allocation diagnostics.

## Verification

- `make validate build check-generated`: passed; full/race tests, coverage,
  script/module gates, sqlc and schema generation checks.
- `GOTOOLCHAIN=go1.25.0 make test-cli CLI_TEST_RUN=TestJSON_`: passed.
- `make bench-cli`: all 84 case-runs pass over three complete runs, retaining
  all 8,400 samples and 39 individual p90-target misses. Every p95/p99/maximum
  guard passes. Worst help/version p90: **3.833 ms**; query p90: **12.900 ms**.
  The 1,000-task JSON tree p90 values are 12.513, 11.718 and 12.900 ms.
- Exact binary SHA-256:
  `3f517c5af4f7609dea572a4a87be1e1136c3e372441c7ea7ba863669c27f1a5e`.
- [Complete reference matrix](review-pr4-r3-latency.json.gz) and
  [receipt/source hashes](review-pr4-r3.json).

The round 2 failures and unchanged control remain failures in their original
receipts. This accepted matrix measures changed source with a demonstrated
allocation reduction; it is not a retry of the failed binary. Native console and
hosted release proof remain Feature 006 obligations. Hosted review of the new
commit proceeds after publication and is not implied by these local checks.

## Third-round approach assessment

The new review contained two suggestions about hypothetical failure diagnostics
in test-only helpers. Both roots already had two recorded rounds. The ce-pov
approach assessment compared another test-only change with retaining the current
bounded, correctly failing tests. It chose the latter: neither suggestion
identifies false success or a present runtime defect. This freezes further churn
on those roots unless observed failures or requirements change. Evaluation was
sequential in the main thread under AGENTS.md; no independent model panel ran.

- [4136958902](https://github.com/newbpydev/tusk/pull/4#discussion_r4136958902): **declined**. The outer-deadline case still fails correctly and reports the actual terminating signal (SIGKILL); it cannot pass as the requested SIGINT/SIGTERM. The five-second escalation assertion is separately bounded, and the four-second startup regression passes. A more specific outer-budget message would improve a hypothetical slow-run diagnosis, but it does not repair a remaining signal-handling or false-success defect. After the third-round approach assessment, keeping the current bounded test instead of another test-only change. Revisit if an observed runner timeout demonstrates that this diagnostic prevents identifying the failure.
- [4136958909](https://github.com/newbpydev/tusk/pull/4#discussion_r4136958909): **declined**. The mode list is hard-coded, all nine current modes are covered, and unknown modes now fail explicitly rather than silently collecting baseline data. The diagnostic report is written only on successful completion by design; it is not the retained-sample reference acceptance report. Up-front validation would improve a future developer-typo failure, but no current requested mode is unknown. After the third-round approach assessment, retaining the independent oracle and explicit failure rather than expanding this diagnostic helper again. Revisit when modes become externally configurable or partial diagnostic recovery becomes a requirement.
