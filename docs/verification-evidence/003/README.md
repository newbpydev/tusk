# Feature 003 execution evidence

U1: [receipt](u1.json) records two red runs and the passing canonical gate. Compressed logs retain raw output. Source hashes identify precommit verified contents; the enclosing unit commit associates this evidence with the accepted revision. Runtime-dependent portions of V07/V09/V11/V12 remain owned by later orchestration/integration units.

U2: [receipt](u2.json) includes calendar/DST, synthetic repeated evening, year-1 regression and RFC UUIDv7 proof. Local parser/identity acceptance; production timezone packaging remains Feature 004.

U3: [receipt](u3.json) covers transaction-local graph/event work, failure rollback and preservation of untouched manual leaf progress.

U6: [receipt](u6.json) proves creation and metadata/manual patch primitives, including a loaded-hierarchy ID collision regression.

U7: [receipt](u7.json) covers public mutations and per-write failure atomicity across five operations.

U4: [receipt](u4.json) proves the complete read facade and suppression of assembled results on cleanup failure.

U5: [receipt](u5.json) maps all 91 local scenarios and retains full/race/coverage/script, Go 1.25, five CGO-free target builds and single-sample benchmark logs. Service coverage is 95.8%, parser 98.4%. The three process barriers verify complete old/new graph/history and fresh integrity/FK/WAL checks. Integration tests characterize existing behavior; canonical target tests supplied the U5 red-first change.

At 10,000 tasks, warm Get/History measured about 0.19/0.26 ms; List/Tree/Stats about 67–72 ms. Full-subtree completion took 1.92 s. Port calls and transaction duration are recorded, not SQL statement traces. These observations do not establish CLI latency or statistical bounds. See the [runbook](../../service.md) for unknown-outcome recovery and pending consumer/native/hosted obligations.

Final [publication review](publication-review.json) and [implementation return](work-return.json) record the seven unit commits, local review scope and independent-review limitation.

PR #3 [review-fix receipt](review-r1.json) records red/green and full canonical verification for UTC year-boundary resolution, including explicit year-zero input rejection and no partial day bounds.
