---
title: Preserve transaction outcomes through error redaction
date: "2026-09-08"
category: database-issues
module: SQLite storage
problem_type: database_issue
component: database
severity: high
symptoms:
  - "Unknown migration cleanup outcome reported only as corrupt data"
  - "Canceled schema inspection reported as corruption or incompatibility"
root_cause: logic_error
resolution_type: code_fix
tags: [sqlite, transactions, error-redaction, cancellation, recovery]
---

# Preserve transaction outcomes through error redaction

## Problem

An error mapper can preserve the reason an operation failed while losing whether its transaction finished. In storage, that distinction changes recovery: a recognizable cause does not prove that repeating the operation is safe.

## Symptoms

The original migration mapper returned a schema category before examining the joined cleanup error. Its regression failed with `unknown cleanup outcome hidden by storage: corrupt data`. Ledger inspection and canonical schema replay also converted cancellation into schema refusal. The historical failures and their fixes are recorded in the [Feature 002 evidence](../../verification-evidence/002/README.md).

## What Didn't Work

Early returns for the most specific schema category looked sufficient because ordinary corruption and incompatibility tests passed. They discarded the independent outcome signal when rollback also failed. Testing each cause alone did not exercise that combination.

Blindly exposing the original error chain is also inappropriate here: the public storage boundary deliberately withholds driver errors, which can contain paths or SQL data. The fix needs to preserve decision-relevant information without restoring that exposure.

## Solution

Map the safe cause first, then attach transaction uncertainty before returning. Preserve cancellation before interpreting a schema operation's failure as schema damage. The implementation in `internal/storage/errors.go:68` follows this order:

```go
cause := storageCause(err)
if errors.Is(err, errIncompatibleSchema) {
    cause = schemaFailure(err, ports.ErrIncompatibleSchema)
} else if errors.Is(err, errCorruptSchema) {
    cause = schemaFailure(err, ports.ErrCorrupt)
}
if errors.Is(err, errMigrationOutcome) {
    return ports.NewTransactionError("migration", cause)
}
return cause
```

The public transaction error exposes `Outcome()` and matches its sanitized cause through `Is`; it has no driver-error `Unwrap` path (`internal/ports/errors.go:24`). Callers can detect both uncertainty and a familiar cause without receiving the original driver error.

## Why This Works

Cause and outcome answer different questions. Cancellation explains why work stopped; it cannot establish whether a commit completed. Corruption describes the inspected data; it cannot establish whether cleanup succeeded. Preserve both facts until the caller chooses recovery.

Keep this policy at the outer storage boundary. A generic mapper that selects a single sentinel cannot also represent every joined transaction state. Conversely, wrapping raw errors solely to preserve state defeats the redaction contract. A safe category plus a typed outcome is enough for this repository.

## Prevention

- Test combinations of cause and cleanup outcome. `TestOpenCause_RetainsMigrationUnknownOutcome` covers corruption, incompatibility, cancellation and deadline causes joined with migration uncertainty in `internal/storage/inspection_cancellation_test.go:50`.
- Assert both `errors.Is` and the typed `Outcome()` result. Checking only the error message or only the cause misses half the contract.
- Cover every declared safe callback sentinel, not a representative subset. Hosted review of [PR #2](https://github.com/newbpydev/tusk/pull/2#discussion_r3963401314) exposed six omissions in the sanitizer; the expanded `TestTransaction_RollbackFailurePreservesCause` checks all current core/port categories plus cancellation and deadlines, with private wrapper text that must stay redacted.
- Keep canceled ledger and canonical-replay tests alongside malformed-schema tests, so schema classification cannot swallow cancellation again.
- Use disk readback after injected commit-before, commit-after and rollback failures. `TestRecovery_AtomicStateAfterFailure` verifies complete old or new task/history state in `internal/storage/recovery_test.go:82`.
- Do not replay a write callback automatically after an unknown outcome. Follow the [storage recovery guidance](../../storage.md).

The fixes and these regressions pass the local `make validate` gate. This evidence does not claim native-platform release validation or protection against power loss under every synchronization policy.

## Related Issues

- [Original review and red/green evidence](../../verification-evidence/002/README.md)
- [Fresh publication review](../../verification-evidence/002/publication-review.json)
