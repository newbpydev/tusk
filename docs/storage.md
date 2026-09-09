# SQLite storage

Feature 002 provides an internal Go repository. The CLI and service have not yet been wired to it. Importing storage, running `tusk --help`, or running `tusk --version` does not open a database.

## Open and close

Call `storage.Open(ctx, storage.Options{Path: path})` and close the returned repository when its owner finishes. An empty explicit path selects `TUSK_DB_PATH`, then an absolute `XDG_DATA_HOME/tusk/tusk.db`, then `~/.local/share/tusk/tusk.db`. Relative overrides resolve against the current working directory. Paths are literal: no shell expansion, globbing or SQLite URI interpretation. An invalid selected override returns an error instead of falling back.

New directories and files use POSIX modes 0700 and 0600; existing permissions are preserved. The final database target must be a regular file and cannot be a symlink or Windows reparse point. Ancestor symlinks are allowed. Use a trusted directory whose contents cannot be replaced by another user. Windows permissions inherit the directory's ACL; native Windows behavior remains a Phase 6 verification gate.

Every physical connection enables foreign keys, a 5000-ms busy timeout and synchronous NORMAL. Disk databases use WAL and SQLite's default automatic checkpointing. One writer connection starts IMMEDIATE transactions; up to four query-only readers use deferred transactions. Contended cancellable writer acquisition checks context between at most 25-ms busy waits, within the original five-second total budget. Application callbacks, statements and commits are never replayed. Native file opening, filesystem I/O and a stuck callback are outside that cancellation bound.

`Close` stops admissions, waits for admitted callbacks and releases both pools. It is idempotent. Do not call it from one of the repository's callbacks: that callback must finish before its owner can close.

## Read and write transactions

Use convenience reads for a single snapshot, or `WithRead` to share one snapshot across multiple queries. Use `WithWrite` for a change set:

```go
err := repo.WithWrite(ctx, func(txctx context.Context, writer ports.TaskWriter) error {
    task, err := writer.GetByID(txctx, id)
    if err != nil {
        return err
    }
    task.Title = title // Validate the proposed business mutation in the service.
    task.UpdatedAt = changedAt
    if err := writer.Update(txctx, task); err != nil {
        return err
    }
    _, err = writer.AppendEvent(txctx, ports.TaskEvent{
        TaskID: id, Kind: ports.EventMetadata,
        ChangedFields: []string{"title"}, OccurredAt: changedAt,
    })
    return err
})
```

Use the callback context (or a derived context) and handle for every operation. Never reenter the repository from a callback. Reentry with its context is rejected; substituting a fresh context evades detection and can deadlock or read a different snapshot. Concurrent handle use and use after callback return are rejected. A read callback does not expose write methods. The first failed write-handle operation invalidates the whole change set, including an error the callback ignores. A callback error or panic rolls back; a panic is rethrown. Returned records are detached; failed reads expose no partial collection.

Storage validates canonical records and representation, not complete business transitions. It preserves immutable creation time, nullable dates/parent, normalized sorted tags, exact notes, and fixed nine-digit UTC timestamps. Tree reads reject corrupt cycles, orphans and depth overflow. `ListChildren` reads immediate children; `GetSubtree` includes the requested task; ancestors are nearest-parent first. List ordering is priority descending, due ascending with nulls last, creation ascending, then ID. Filtering finishes through the core oracle after a bound SQL candidate query.

History is metadata-only. Storage appends an event only when explicitly requested, assigns increasing sequences, and deletes events with their task. Nonrecursive deletion refuses tasks with children; recursive deletion reports the sorted exact deleted IDs. The Phase 3 service must own cycle/depth validation before writes, legal status changes, moves, rollups, freshness checks, event selection and atomic orchestration. Do not treat adapter acceptance as service acceptance.

## Refusal and recovery

Open inspects the database in a read-only snapshot before changing its journal mode. Recognized stores carry Tusk's application ID, the exact applied migration prefix, checksums and canonical schema catalog. Pending forward migrations and ledger entries commit together. Foreign databases, newer schemas, missing/drifted objects and mismatched ledgers are refused. Do not modify a shipped migration, delete sidecars, run a down migration or reset a refused database. Preserve the files and use a compatible binary or investigate a disposable copy.

Use `errors.Is` for core errors, context cancellation and `ports` storage categories. Driver errors and raw paths/notes/SQL are sanitized at the public boundary. A `ports.TransactionError` has `Outcome() == "unknown"`: a commit or cleanup acknowledgment could not be confirmed. It preserves safe cause matching but cannot tell the caller whether the change persisted. Close the failed owner and read back tasks and events through a fresh repository before deciding on another action. Never blindly rerun a callback after this error.

WAL with synchronous NORMAL preserves transactional consistency in the tested process-death cases. The crash tests do not simulate hardware power loss or promise that every acknowledged commit survives it. A long read snapshot can delay checkpoint progress and grow the WAL; keep callbacks short. The lifecycle test proves automatic WAL restart/recycling after releasing the reader, without issuing a test checkpoint.

## Consistent offline backup

1. Stop every process using this database and successfully close every repository. If an owner failed to close or another process is active, do not assume a consistent offline state.
2. Preserve the original database and any remaining `-wal` and `-shm` sidecars together. Do not copy only the main file while WAL writers or readers are active. For a normally closed store, SQLite checkpoints the WAL and removes those sidecars; the offline-backup replay tests that case.
3. Copy into a private directory, keeping matching sidecar names if any remain. Open the copy with the same compatible binary; confirm expected tasks/history and integrity before using the backup. Keep the original unchanged.

There is no production backup, repair, reset or inspection CLI in this phase. Integrity fixtures use `PRAGMA integrity_check` (one `ok` result) and `PRAGMA foreign_key_check` (zero rows) through test-owned connections.

## Reproduce local evidence

Migration filenames use contiguous, fixed-width prefixes from `001` through `999`
so lexicographic and numeric order agree. Expanding this inventory requires an
explicit format change. The conservative authoring check scans the entire SQL
file, including comments and string literals, for transaction/file-state keywords
(`begin`, `commit`, `end`, `rollback`, `savepoint`, `release`, `attach`, `detach`,
`vacuum`, `pragma`). Avoid those words even in commentary or seed data; this check
rejects such files rather than parsing SQL tokens.

Run all operations through Make from the repository root:

```sh
make setup
make validate
make check-generated
GOTOOLCHAIN=go1.25.0 make test-compat build-storage
make bench-storage
```

`check-generated` needs the pinned tool installed by `make setup-sqlc`; ordinary tests and plain `make` do not download it. Generation owns `internal/storage/sqlc`, uses configuration version 2 and sqlc v1.31.1, and compares complete directories without requiring Git metadata. Use `make generate` after editing query inputs. Never edit generated files manually.

`make validate` runs the runbook fixtures along with the full/race/coverage/script gates:

| Behavior | Executable evidence |
| --- | --- |
| Recognized/current, foreign/newer/drifted and failed initialization | `TestMigrate_RefusesForeignOrNewer`, `TestMigrate_BrokenLedger`, `TestOpen_FailureStagesReleaseEveryHandle`, migration fault tests |
| Task/event readback after unknown outcomes | `TestRecovery_AtomicStateAfterFailure` |
| Two-process snapshots, contention and cancellation | `TestDiskReaders_SnapshotAcrossWriterCommit`, `TestDiskWriter_BusyAndCancelRecovery`, `TestDiskWriters_NoLostUpdate` |
| Reaped process death before/after commit | `TestRecovery_KilledWriterAtomic` |
| Integrity, automatic checkpoints and lock-free cleanup | `TestDiskLifecycle_IntegrityAndCleanup` |
| Normally closed offline backup | `TestStorageRunbookReplay_OfflineBackup` |

All database fixtures are temporary. Shared-memory fixtures establish semantics and lifetime only; real disk/process fixtures establish WAL behavior. `make build-storage` cross-compiles storage and tests for Linux amd64/arm64, macOS amd64/arm64 and Windows amd64 with CGO disabled. Cross-compilation is not native execution. Hosted checks, native Windows/macOS runtime, CLI latency and final service integration remain separate later gates.

Storage benchmark fixtures use 0/100/1000/10000 tasks, 128-byte notes, three tags and ten-task chains of depth ten. Population and cleanup are outside measured loops. Current-open includes close and first-create includes initialization plus close. These storage measurements do not prove the end-to-end CLI's 15-ms budget. Raw measurements and machine details are recorded in the paired Feature 002 verification evidence.
