# Feature 002 local evidence

Reference environment: Go 1.27.1-X:nodwarf5 and explicit Go 1.25.0; Linux 7.2.3-1-cachyos, amd64; AMD Ryzen 5 4500U (six benchmark workers). Modernc SQLite v1.58.0 / libc v1.75.6 reports SQLite 3.53.4. Normal temporary test fixtures use tmpfs-backed regular files; they exercise actual SQLite WAL and separate processes. The final benchmark uses the workspace's Btrfs filesystem through an ignored private `bin/storage-bench-tmp` directory.

## Commit reconstruction

The original red-first work accumulated without unit commits. The user required a separate validated commit before every next unit. Each completed unit was reconstructed in isolation, validated again on its own contents, then committed before the next reconstruction. The reconstructed code at U4 was byte-compared with the preserved implementation before advancing the feature branch. Only U5 test files remained uncommitted afterward.

| Unit | Commit | Gate on that unit's contents | Storage coverage |
| --- | --- | --- | --- |
| U6 | `ca29a9b` | `make validate` | 97.8% |
| U1 | `4122b17` | `make validate` | 96.1% |
| U2 | `938e7f4` | `make validate check-generated` | 96.1% |
| U3 | `a9e21d6` | `make validate check-generated` | 95.7% |
| U4 | `017d599` | `make validate check-generated` | 97.0% |
| U5 | `63d6691` | `make validate check-generated build` | 97.6% |

Original red/green receipts remain in the paired verification plan as historical evidence; their references to then-uncommitted work are not current branch status.

## Benchmarks

[Raw output](storage-benchmarks.txt), with trailing whitespace trimmed, comes from `TMPDIR=<repo>/bin/storage-bench-tmp make bench-storage`, with fixture setup outside timed loops, 128-byte notes, three tags, and groups of depth-ten chains. The 1,000-task sample measures GetByID 0.058 ms, List 7.30 ms, ten-task GetSubtree 0.312 ms and OpenCurrent including close 2.23 ms. FirstCreate including close measures 22.1 ms; the 10,000-task full-list stress sample takes 66.2 ms and approximately 24.8 MB allocated per operation. These are diagnostic samples, not latency thresholds or whole-CLI proof. Initial exploratory runs used different fixture shapes and are not the canonical benchmark record.

## Code review

`ce-simplify-code` ran its reuse, quality and efficiency passes sequentially under the user-supplied AGENTS tool mapping. No behavior-preserving edit was warranted; safety checks and canonical core operations were retained.

`ce-code-review` reviewed base `4bbc639` through U4 plus U5. The [original report](code-review.json) preserves two actionable findings and its completed review receipt. Correctness, standards, testing, maintainability, security, performance, API contract, migration and reliability passes ran in the parent context, so they do not count as independent model reviews.

### Actionable Findings

Both reported findings were reproduced in a faithful isolated copy and again in the feature checkout before fixes:

1. Schema errors masked unknown migration cleanup outcomes. `TestOpenCause_RetainsMigrationUnknownOutcome` failed with `unknown cleanup outcome hidden by storage: corrupt data`. The public mapper now retains the safe cause inside `TransactionError`.
2. Ledger inspection and canonical replay masked cancellation as corruption/incompatibility. `TestInspection_PreservesCancellationCause` failed on both paths. Shared schema error mapping now preserves cancellation/deadline categories while retaining actual schema refusal categories.

The caller applied both findings and reran focused migration/inspection tests successfully. No actionable review finding remains unapplied.

### Coverage

Independent cross-model review was unavailable. The Claude route requested claude-opus-5 at high effort but received a provider 401 expired-OAuth response; the one replacement route requested composer-2.5-fast through Cursor but its CLI rejected `--add-dir` before review. Neither produced a usable review or served-model receipt. All supervised jobs terminated and their consumed job directories were removed. Local passes and scratch reproductions remain attributed evidence, not independent corroboration.

Native Windows/macOS runtime, hosted checks and CLI latency remain unexecuted. Current/minimum compiler runs and cross-compilation do not substitute for them. The Windows path/reparse scenario stays unchecked in the feature verification plan.

### Verdict

The review's two required fixes are applied with red/green evidence. Final aggregate acceptance and its code fingerprint are recorded in the paired verification plan and `acceptance.json`; no push, PR, merge or release is implied.

## Publication review — 2026-09-08

The user subsequently authorized simplify, review, compound and commit/push/PR. The [fresh review](publication-review.json) covers origin/main through U5, including both earlier fixes: zero actionable findings. Reuse, quality and efficiency passes found no worthwhile changes. make validate check-generated build passes (storage 97.6%). Both external routes again failed without usable review evidence; local passes are not independent corroboration. Publication state is tracked in MASTERPLAN target 2.4.

The [compounded learning](../../solutions/database-issues/preserve-transaction-outcomes-through-error-redaction.md) explains why safe error causes and unknown outcomes must be preserved separately. Its frontmatter and claims validators pass; source grounding ran sequentially in the parent context. The task-event glossary now reflects implemented persistence.

Published as [PR #2](https://github.com/newbpydev/tusk/pull/2) against main. Local acceptance and its original fingerprint remain historical evidence; publication does not change native/hosted release gates or imply merge.
