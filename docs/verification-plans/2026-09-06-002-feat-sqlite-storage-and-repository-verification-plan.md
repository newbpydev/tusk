---
feature-id: "002"
plan-source: docs/plans/2026-09-06-002-feat-sqlite-storage-and-repository-plan.md
surface-profiles: [library-sdk, data-persistence-migration, infrastructure-operations]
status: locally accepted - native and hosted release gates deferred
evidence-scope: local implementation, per-unit commits and acceptance; no native or hosted execution
---

# Feature 002 Verification Plan

## Verification contract

This document accompanies the [implementation plan](../plans/2026-09-06-002-feat-sqlite-storage-and-repository-plan.md) and [issue workorder](../workorders/2026-09-06-002-feat-sqlite-storage-and-repository-issues-workorder.md). It covers 22 requirements through 68 planned scenarios across six units, executed in order U6, U1, U2, U3, U4, U5.

Test the repository boundary, canonical schema, generated queries, explicit file opening, transaction outcomes, and actual disk recovery. Service business orchestration, CLI/TUI workflows, hosted CI, and publication belong to later plans. No test, dependency installation, generation, or application build was performed during planning.

The implementation's minimum compiler is Go 1.25; its exact engine/graph must match KTD1. Local Linux execution and CGO-disabled cross-builds are Phase 2 evidence. Native Windows/macOS runtime proof remains pending until performed under Phase 6; scenario 002-V33 supplies that reusable native test contract.

## Requirement coverage

| Requirement | Units | Scenarios | Evidence tier |
| --- | --- | --- | --- |
| R1 | U6, U5 | 002-V01, 002-V02, 002-V07, 002-V67 | Local + cross-build |
| R2 | U6, U3 | 002-V03, 002-V29 | Local focused/full/race |
| R3 | U3 | 002-V29, 002-V30, 002-V31, 002-V33, 002-V35 | Local focused/full/race |
| R4 | U3 | 002-V30, 002-V31, 002-V32, 002-V33, 002-V35 | Local + native Windows deferred |
| R5 | U6, U3, U4, U5 | 002-V03, 002-V06, 002-V08, 002-V34, 002-V36, 002-V37, 002-V38, 002-V57, 002-V65 | Local focused/full/race |
| R6 | U1, U3, U5 | 002-V09, 002-V10, 002-V11, 002-V12, 002-V13, 002-V14, 002-V18, 002-V39, 002-V65, 002-V68 | Local focused/full/race |
| R7 | U1, U3, U5 | 002-V09, 002-V10, 002-V11, 002-V12, 002-V13, 002-V14, 002-V15, 002-V16, 002-V17, 002-V18, 002-V20, 002-V34, 002-V39, 002-V62, 002-V63, 002-V64, 002-V65, 002-V68 | Local focused/full/race |
| R8 | U1, U2, U4 | 002-V09, 002-V19, 002-V21, 002-V40, 002-V41, 002-V42 | Local focused/full/race |
| R9 | U1, U4 | 002-V14, 002-V40, 002-V41, 002-V42, 002-V47, 002-V58 | Local focused/full/race |
| R10 | U1, U2, U4 | 002-V09, 002-V19, 002-V21, 002-V40, 002-V48, 002-V49 | Local focused/full/race |
| R11 | U4 | 002-V40, 002-V42, 002-V43, 002-V46, 002-V48, 002-V49, 002-V50 | Local focused/full/race |
| R12 | U2, U4 | 002-V23, 002-V24, 002-V44, 002-V45 | Local focused/full/race |
| R13 | U2, U4 | 002-V22, 002-V46, 002-V47 | Local focused/full/race |
| R14 | U6, U3, U4, U5 | 002-V04, 002-V05, 002-V08, 002-V36, 002-V51, 002-V52, 002-V54, 002-V55, 002-V56, 002-V59, 002-V60, 002-V62, 002-V63, 002-V64 | Local focused/full/race |
| R15 | U3, U4 | 002-V38, 002-V53, 002-V54, 002-V57 | Local focused/full/race |
| R16 | U6, U1, U3, U4, U5 | 002-V06, 002-V16, 002-V17, 002-V34, 002-V35, 002-V38, 002-V50, 002-V52, 002-V53, 002-V54, 002-V55, 002-V56, 002-V57, 002-V58, 002-V61, 002-V64, 002-V68 | Local focused/full/race |
| R17 | U6, U1, U4 | 002-V04, 002-V19, 002-V41, 002-V47, 002-V49, 002-V50 | Local focused/full/race |
| R18 | U1, U2, U5 | 002-V12, 002-V20, 002-V21, 002-V22, 002-V23, 002-V24, 002-V25, 002-V26, 002-V27, 002-V28, 002-V67 | Local focused/full/race |
| R19 | U3, U5 | 002-V37, 002-V60, 002-V62, 002-V63, 002-V65 | Local focused/full/race |
| R20 | U6, U1, U3, U4, U5 | 002-V05, 002-V06, 002-V08, 002-V18, 002-V36, 002-V51, 002-V56, 002-V59, 002-V60, 002-V61, 002-V62, 002-V63, 002-V64, 002-V65 | Local focused/full/race |
| R21 | U6, U2, U5 | 002-V01, 002-V02, 002-V26, 002-V28, 002-V67, 002-V68 | Local focused/full/race |
| R22 | U6, U3, U5 | 002-V07, 002-V33, 002-V66, 002-V67, 002-V68 | Local + later native/hosted |

## Scenarios

Checkboxes record accepted scenarios after the owning unit's canonical gate. U6 is accepted after the canonical gate; the execution record retains its original failures and resolution. Test names outside U6 remain planned identifiers. The unit's Files field owns the test location.

### U6. Runtime compatibility

002-ISS-021 resolution adds cancellation without a deadline, lock release during acquisition, bounded busy exhaustion, and timeout restoration after success/failure. Inject configuration/restore/rollback failures to prove poisoned connections are discarded. The 100-ms regression remains unchanged. Only failed driver BeginTx before a callback may be retried; no statement/callback replay is permitted.

- [x] 002-V01 **Compatibility — TestSQLiteCompatibility** (covers R1, R21). Inspect the resolved module graph and query sqlite_version() through the selected driver under an explicitly selected Go 1.25 compiler. **Expect:** Require sqlite v1.58.0, libc v1.75.6, SQLite 3.53.4; a different graph or engine fails. Record compiler identity separately from the installed Go 1.27 result.

- [x] 002-V02 **Boundary — TestSetupGoMinimum** (covers R1, R21). Use fake Go executables reporting missing, malformed, 1.24, 1.25, and newer versions in the setup script fixture. **Expect:** Missing, malformed, and 1.24 fail clearly; 1.25 and newer pass. Setup must not download a compiler or dependency.

- [x] 002-V03 **Purity — TestConnectorConstruction_NoIO** (covers R2, R5). Construct two connectors using literal temporary paths without opening or pinging them. **Expect:** No file/directory appears and no network call occurs; the connectors carry independent immutable configuration.

- [x] 002-V04 **Protection — TestReaderConnection_QueryOnly** (covers R14, R17). Open a reader physical connection and begin a deferred read transaction; attempt a write through a test-only raw handle. **Expect:** query_only rejects mutation, while normal reads work. ReadOnly transaction options alone are not accepted as the protection.

- [x] 002-V05 **Locking — TestWriterConnection_BeginsImmediate** (covers R14, R20). Hold an IMMEDIATE transaction on writer A; begin writer B before any application read using a bounded deadline. **Expect:** B cannot enter its callback until A releases the writer lock. A deferred transaction that reaches the callback prematurely fails this test.

- [x] 002-V06 **Cancellation — TestAcquire_CanceledContext** (covers R5, R16, R20). Cancel before pool acquisition, then occupy a single writer and cancel a waiting second acquisition; separately hold a SQLite writer lock from another connection and use a 100-ms deadline. **Expect:** Canceled acquisitions invoke no callback; the SQLite lock wait returns the context error before the 5000-ms busy limit. A later acquisition succeeds. Failure blocks U1 until compatibility behavior is resolved.

- [x] 002-V07 **Cross-build — BuildStorageTargets** (covers R1, R22). Use make build-storage to compile storage with CGO disabled for linux amd64/arm64, darwin amd64/arm64, windows amd64. **Expect:** Every target builds; capture exact compiler and dependencies. Cross-built output is not native execution evidence.

- [x] 002-V08 **Replacement — TestConnectionFactory_Pragmas** (covers R5, R14, R20). Retire and reopen physical writer/reader connections through the private factory. **Expect:** Each replacement has foreign_keys=1, busy_timeout=5000, synchronous=NORMAL and the correct query_only/begin mode; no process-global configuration changes.

### U1. Schema and migrations

- [x] 002-V09 **Normal — TestMigrate_EmptyDatabase** (covers R6, R7, R8, R10). Migrate a new disk file and a unique memory fixture. **Expect:** Expect application ID, three application tables, required indexes/constraints, one matching ledger entry and no fixture-only SQL.

- [x] 002-V10 **Idempotency — TestMigrate_ReopenDoesNothing** (covers R6, R7). Populate a migrated fixture, capture logical schema/ledger/tasks/events, and open repeatedly. **Expect:** The migration callback executes zero already-applied scripts; all captured values and applied-at times remain equal.

- [x] 002-V11 **Malformed inventory — TestMigrationInventory_Invalid** (covers R6, R7). Inject filesystems with duplicate numeric prefixes, missing 001, a gap, invalid suffix, empty SQL, and forbidden transaction-control migration content. **Expect:** Validation fails before schema writes; report the offending bounded filename/version without user data.

- [x] 002-V12 **Compatibility — TestMigrationInventory_StableBytes** (covers R6, R7, R18). Compare canonical embedded SQL bytes with LF-controlled source and change one applied script byte in a fixture. **Expect:** Canonical checkout/builds preserve SHA-256 across platforms; a changed applied digest refuses open without modifying the existing database.

- [x] 002-V13 **Refusal — TestMigrate_RefusesForeignOrNewer** (covers R6, R7). Open fixtures with another application ID, unbranded user tables, and a newer well-formed Tusk ledger using read-only inspection. **Expect:** Fail before intentional journal-mode/schema/data writes; compare main/WAL hashes and logical data, excluding transient SHM locks. Never initialize over foreign data.

- [x] 002-V14 **Corruption — TestMigrate_BrokenLedger** (covers R6, R7, R9). Supply a branded DB with missing/malformed ledger, non-prefix versions, conflicting filenames, invalid hashes, or missing required schema objects. **Expect:** Return corruption/incompatible-schema as specified; preserve files and never accept a partial ledger as fresh initialization.

- [x] 002-V15 **Partial failure — TestMigrate_FailurePreservesPreviousVersion** (covers R7). Start at schema 001 with sentinel rows and inject fixture 002 that changes data then fails on its next statement. **Expect:** Reopen shows exactly schema 001, old rows/events and old ledger; no partial column/index/data change survives.

- [x] 002-V16 **Ledger failure — TestMigrate_LedgerInsertRollback** (covers R7, R16). Let pending DDL succeed then fail its ledger insert through the private test connector. **Expect:** Rollback restores prior schema/data and returns failure; the next correct migration attempt succeeds once.

- [x] 002-V17 **Commit failure — TestMigrate_CommitOutcome** (covers R7, R16). Inject commit failure before forwarding and an acknowledgment failure after a real commit. **Expect:** Both report no confirmed success; outcome is unknown where needed. Reopen establishes old-or-new atomic state, never a mixed ledger/schema.

- [x] 002-V18 **Concurrency — TestMigrate_ConcurrentInitializers** (covers R6, R7, R20). Use two processes and handshakes to initialize the same new file or advance the same old fixture. **Expect:** Exactly one applies each migration; the second rechecks the ledger under its lock. An observed busy failure is bounded and leaves a valid database.

- [x] 002-V19 **Constraints — TestSchema_RejectsInvalidRows** (covers R8, R10, R17). Attempt raw inserts with null/duplicate IDs, empty/long title, invalid status/priority/progress, invalid JSON shape, self/missing parent, and inconsistent completion. **Expect:** CHECK/FK/NOT NULL reject local violations. Open-parent progress 100 is permitted; cross-row cycle/rollup prevention is not attributed to these constraints.

- [x] 002-V20 **Fixture isolation — TestEmbed_ForwardOnly** (covers R7, R18). Read the embedded inventory and sqlc schema inputs, then exercise the disposable inverse fixture on a temporary DB. **Expect:** Only forward production migrations are embedded/generated; inverse fixture cleans its disposable schema and is never used for installed-data recovery.

### U2. Generated queries and tooling

- [x] 002-V21 **Generated CRUD — TestQueries_CRUDAndCounts** (covers R8, R10, R18). Use generated create/get/update/delete and event queries against the real migration schema, including absent and duplicate rows. **Expect:** Correct fields/counts/NULL values and sequence are returned; adapter error mapping is tested separately in U4.

- [x] 002-V22 **Generated traversal — TestQueries_TreeSets** (covers R13, R18). Seed root, siblings, depth-10 branch and unrelated root; query children/subtree/ancestors, then inject a cycle in a fixture. **Expect:** ID sets match the expected membership; distinct-ID recursion terminates on cycles. Adapter ordering/corruption checks remain U4.

- [x] 002-V23 **Candidate parity — TestQueries_CandidateSuperset** (covers R12, R18). Enumerate status/priority/date/parent combinations over a fixed corpus with mixed valid/invalid filter values. **Expect:** Every task selected by core survives SQL candidate selection. The adapter's final core pass determines exact result/order.

- [x] 002-V24 **Injection — TestQueries_BoundValues** (covers R12, R18). Pass quotes, semicolons, percent/underscore, comment syntax, and Unicode through task IDs and filter parameters. **Expect:** Values remain data; no table changes, broader predicate, network access, or extra statement execution occurs.

- [x] 002-V25 **Tool provenance — TestSQLCTool_RejectsInvalidAsset** (covers R18). Use local archive/downloader doubles for wrong version/digest, truncated download, unexpected executable, traversal member and unsupported host. **Expect:** setup-sqlc fails before installation/execution; prior tool remains usable. The real pinned archive is validated before install during implementation.

- [x] 002-V26 **Reproducibility — TestCheckGenerated_DetectsDrift** (covers R18, R21). Generate twice from the same input, then alter/add/remove one output file in a fixture copy. **Expect:** Identical generation matches byte-for-byte; stale, extra, and absent outputs fail. check-generated leaves its input tree unchanged.

- [x] 002-V27 **Recovery — TestGenerate_FailurePreservesOutput** (covers R18). Make sqlc fail on malformed SQL or after producing only part of an output directory. **Expect:** No checked-in output is replaced until generation succeeds; old outputs and source SQL survive.

- [x] 002-V28 **Source archive — TestCheckGenerated_WithoutGit** (covers R18, R21). Run generation/check in a fixture without .git and in a tree where generated files are untracked. **Expect:** Comparison still detects drift; it cannot depend on git diff ignoring untracked files. make test does not implicitly download sqlc.

### U3. Open and file lifecycle

- [x] 002-V29 **Configuration — TestResolvePath_Precedence** (covers R2, R3). Inject override/XDG/home/cwd combinations, including empty override, relative override, relative XDG, missing home and missing cwd when needed. **Expect:** Follow exact precedence and relative-path contract; irrelevant failed fallback lookups do not prevent a valid explicit path.

- [x] 002-V30 **Literal paths — TestOpen_LiteralFilename** (covers R3, R4). Use spaces, Unicode, percent, #, ?, = and a filename beginning file: or equal to :memory: where the host permits. **Expect:** Create only the literal intended file, never a URI-selected memory DB or alternate mode. Host-invalid names return a path error.

- [x] 002-V31 **Invalid targets — TestOpen_RejectsUnsafeTarget** (covers R3, R4). Try empty resolved path, invalid UTF-8, NUL, directory, FIFO/non-regular target, final symlink, and a read-only invalid path. **Expect:** Refuse before DB use and never select another path. Ancestor symlinks that resolve to legitimate directories remain permitted.

- [x] 002-V32 **POSIX permissions — TestOpen_POSIXPermissions** (covers R4). With restrictive and permissive umasks, create a new app directory/file; also open an existing regular file and symlinked ancestor. **Expect:** New permissions are no broader than 0700/0600; existing directory/file modes are preserved. Privileged runners use injected denial rather than claim mode-bit denial.

- [ ] 002-V33 **Native Windows — TestOpen_WindowsPathsAndReparse** (covers R3, R4, R22). On Windows, exercise drive paths, spaces/Unicode, reserved filenames, final reparse targets and inherited private-profile ACLs. **Expect:** Literal path policy and unsafe-target refusal hold. Linux and cross-build results leave this native acceptance record pending for Phase 6.

- [x] 002-V34 **Open failure — TestOpen_ClosesPartialResources** (covers R5, R7, R16). Fail each directory/file/compatibility/WAL/migration/writer/reader stage via private filesystem or connector seams. **Expect:** All opened handles close, existing files and sidecars remain, primary error stays visible, and a subsequent valid Open succeeds.

- [x] 002-V35 **No fallback — TestOpen_InvalidOverrideNeverFallsBack** (covers R3, R4, R16). Set an explicit unwritable/invalid path while valid XDG/home alternatives exist. **Expect:** Return the override error and create no fallback DB/directories. Diagnostics contain no note contents or raw DSN.

- [x] 002-V36 **Pool integration — TestOpen_ReplacementConnectionPragmas** (covers R5, R14, R20). Open disk storage, force writer/read physical replacement, inspect pool bounds and connection settings. **Expect:** Writer max-open=1, readers max-open=4; each replacement remains configured and readers stay query-only.

- [x] 002-V37 **Memory isolation — TestOpen_MemoryLifetimeAndIsolation** (covers R5, R19). Run two uniquely named memory fixtures in parallel, close readers, replace a reader, and finally close the writer anchor. **Expect:** Fixtures never share rows; MEMORY journal is observed; data persists until all fixture handles close. No global name counter/state is required.

- [x] 002-V38 **Close lifecycle — TestOpen_CloseAdmissionAndIdempotency** (covers R5, R15, R16). Hold an admitted callback, start Close, attempt a new repository operation, then release callback and repeat Close. **Expect:** Existing callback finishes; new admission fails; all pools close; second Close is safe; post-close calls return closed-repository. Close is never invoked inside a callback.

- [x] 002-V39 **Current schema — TestOpen_CurrentSchemaNoMigrationWrite** (covers R6, R7). Open current and newer schema fixtures while another process has the write lock and inspect compatibility access. **Expect:** Current-schema inspection does not take an unnecessary migration writer transaction; newer refusal does not change journal mode or task/ledger/main-WAL bytes.

### U4. Repository contracts

- [x] 002-V40 **Round trip — TestRepository_RoundTripAndDetachedValues** (covers R8, R9, R10, R11). Round-trip all task fields, zero/empty notes, nil/empty tags, absent/present dates and root/non-root parents. **Expect:** Return canonical detached values with non-nil empty collections; description stays exact and optional NULL differs from empty malformed data.

- [x] 002-V41 **Bad input — TestCodec_RejectsInvalidInput** (covers R8, R9, R17). Try nil task, invalid UTF-8/NUL, noncanonical ID/title/tag set, invalid status/range/time, and changed CreatedAt on Update. **Expect:** Reject without changing stored state; nil tags may encode as []; open parent progress 100 is accepted at the row boundary.

- [x] 002-V42 **Corrupt decode — TestCodec_RejectsCorruptRows** (covers R8, R9, R11). Bypass constraints only in fixtures and store malformed timestamps/JSON/tag elements, completion mismatch or invalid text; separately inject a row-iteration error after one valid row. **Expect:** Read fails with corruption or the mapped iteration error and no partial collection; no default clock, trim, clamp, rewrite, or auto-repair occurs.

- [x] 002-V43 **Ownership — TestRepository_ReturnedValuesDoNotAlias** (covers R11). Mutate every returned slice, tag, pointer, event changed-field slice and task field, then read again. **Expect:** Database and subsequent results are unchanged; inputs cannot mutate committed values after calls return.

- [x] 002-V44 **Filter parity — TestRepository_FilterMatchesCore** (covers R12). Compare ordered IDs against core across statuses/priorities/tags, mixed invalid enum values, conflicting root/parent, exclusive due bounds, Ä search and literal wildcard text. **Expect:** Exact core membership holds; empty filter includes done, and no CLI-only defaults or implicit truncation appear.

- [x] 002-V45 **Ordering — TestRepository_DefaultSort** (covers R12). Seed equal priorities/timestamps, missing due dates, Unicode titles and out-of-insertion-order IDs. **Expect:** Order is priority descending, due ascending/null last, created ascending, ID ascending; siblings use the same order.

- [x] 002-V46 **Hierarchy — TestRepository_TreeReadContract** (covers R11, R13). Read a root, non-root subtree, leaf, root ancestors, depth-10 branch, missing task and existing task with no children. **Expect:** Target-first depth-first subtree and parent-first ancestor ordering match the plan; original ParentID is retained; empty is distinct from missing.

- [x] 002-V47 **Corrupt graph — TestRepository_TraversalRejectsCorruption** (covers R9, R13, R17). Inject self/two-node/deep cycles, an orphan, and a subtree whose stored ancestor depth exceeds 10. **Expect:** Queries terminate under a deadline and return corruption plus matching core sentinel; no orphan promotion or returned partial graph.

- [x] 002-V48 **History — TestRepository_EventRetention** (covers R10, R11). Append each event kind with fixed times, delete the highest committed event's task, append again, and query task history. **Expect:** Sequence ascends without reuse of a deleted committed maximum; gaps are allowed. Events contain field names only, with sorted unique allowlisted fields and no old title/notes.

- [x] 002-V49 **Deletion — TestRepository_DeleteContract** (covers R10, R11, R17). Delete leaf and parent with recursive false/true while an unrelated root and events remain; repeat against a corrupt subtree fixture. **Expect:** False rejects children; true returns sorted exact deleted IDs and cascades their events. Missing task or corrupt traversal fails; unrelated rows/events survive.

- [x] 002-V50 **Error mapping — TestRepository_DomainErrors** (covers R11, R16, R17). Exercise missing read/update/delete/parent/event target, duplicate create, invalid record, busy, read-only, full/I/O and closed handles. **Expect:** errors.Is matches the intended domain/context/port category; SQLite concrete types and raw SQL do not cross the port.

- [x] 002-V51 **Read snapshot — TestWithRead_StableSnapshot** (covers R14, R20). Read a row, commit a concurrent title/history change, and read both again within the same WithRead. **Expect:** Both reads see the original committed snapshot; a new snapshot sees the new pair. The callback has no mutation methods.

- [x] 002-V52 **Atomic rollback — TestWithWrite_ChildAndHistoryRollback** (covers R14, R16). Update a child and ancestor, then fail event append or callback; separately fail rollback using the private connector. **Expect:** Confirmed rollback leaves every row/event unchanged. Unconfirmed rollback returns unknown outcome and discards the connection.

- [x] 002-V53 **Callback misuse — TestTransaction_CallbackLifetime** (covers R15, R16). Try nil callback, nested repository use with callback context, retained handle after exit, concurrent handle calls, and a panic. **Expect:** Nil/nested/escaped/concurrent use fails before unsafe SQL; panic rolls back then re-panics; no reusable active transaction or race remains.

- [x] 002-V54 **Swallowed error — TestWithWrite_LatchesOperationFailure** (covers R14, R15, R16). Succeed one update, force a later writer-handle operation to fail, let the callback ignore that error and return nil. **Expect:** WithWrite still fails and rolls back all prior changes. A swallowed statement error cannot turn partial business work into success.

- [x] 002-V55 **Unknown commit — TestWithWrite_CommitFailureNeverReplays** (covers R14, R16). Inject failure before commit reaches SQLite and failure after SQLite committed but before acknowledgment reaches the adapter. **Expect:** No callback replay; report unknown outcome where proof is absent. Readback shows an atomic old/new state with exactly zero/one event set.

- [x] 002-V56 **Cancellation boundary — TestWithWrite_CancellationAtCommit** (covers R14, R16, R20). Cancel before begin, during a statement, just before commit attempt, and after successful commit acknowledgment. **Expect:** Pre-commit cancellation rolls back when confirmed; known commit success remains success. Do not claim instant cancellation of native open or background-context Commit.

- [x] 002-V57 **Connection recovery — TestTransaction_DiscardsPoisonedConnection** (covers R5, R15, R16). Inject commit/rollback/read-cleanup failure, then issue a fresh operation through the same repository. **Expect:** Failed physical connection is discarded with supported pool APIs; replacement is configured and succeeds; no lingering transaction or lock remains.

- [x] 002-V58 **Error privacy — TestStorageErrors_RedactData** (covers R9, R16). Use sentinel secret-like note/title/path values while provoking decode, SQL, context, close and commit failures. **Expect:** Messages expose only bounded operation/category/metadata; no raw SQL, stored notes, or concrete driver unwrap chain appears. Context/domain matching remains intact.

### U5. Disk acceptance and handoff

- [x] 002-V59 **Two writers — TestDiskWriters_NoLostUpdate** (covers R14, R20). Two repository instances concurrently increment a fixture value by read-modify-write inside WithWrite, synchronized at lock acquisition. **Expect:** All successful increments survive; callbacks serialize and event counts equal acknowledged changes. No lost update or unbounded retry.

- [x] 002-V60 **Two processes — TestDiskReaders_SnapshotAcrossWriterCommit** (covers R14, R19, R20). Separate processes use the same temporary WAL DB, with channel/pipe handshakes marking snapshot acquisition and writer commit. **Expect:** Reader retains old snapshot while writer commits; next read sees new state. Observe actual journal_mode=wal.

- [x] 002-V61 **Lock bounds — TestDiskWriter_BusyAndCancelRecovery** (covers R16, R20). External process holds the write lock; test the 5000-ms busy limit and a 100-ms context, then release the holder. **Expect:** The short-deadline operation matches the context error and returns before the full 5000-ms limit; an uncanceled wait returns busy within the documented scheduling tolerance. A later write succeeds. Failure of the deadline behavior is a compatibility defect, not grounds to weaken this acceptance case.

- [x] 002-V62 **Crash before commit — TestRecovery_KilledWriterBeforeCommit** (covers R7, R14, R19, R20). Helper process changes task/event rows in one transaction and signals a pre-commit checkpoint; terminate it, reap it, reopen DB. **Expect:** Old complete task/event set remains; integrity_check is ok and foreign_key_check has no rows.

- [x] 002-V63 **Crash after commit — TestRecovery_KilledWriterAfterAck** (covers R7, R14, R19, R20). Helper commits task/event changes and signals confirmed commit, then terminate/reap and reopen. **Expect:** Exactly one committed set survives process death. This is not a simulated power-loss durability claim.

- [x] 002-V64 **Recovery readback — TestRecovery_AtomicStateAfterFailure** (covers R7, R14, R16, R20). Repeat injected statement/event/commit/rollback failures on disk and reopen with an independent repository. **Expect:** Only whole old/new transaction states occur as appropriate; unknown outcomes are resolved by inspection, never blind replay.

- [x] 002-V65 **Integrity and cleanup — TestDiskLifecycle_IntegrityAndCleanup** (covers R5, R6, R7, R19, R20). Run repeated opens, migrations, reads/writes, cancellation and process termination; also hold a reader while writes grow WAL, release it, and continue writes beyond the automatic checkpoint threshold. **Expect:** Integrity/FK checks pass, checkpoint progress or bounded subsequent WAL recycling is observed, children are reaped, and no live locks block cleanup. File shrinking alone is not the checkpoint oracle, and an explicit test checkpoint cannot be reported as automatic-checkpoint proof. No real developer database is touched.

- [x] 002-V66 **Performance — BenchmarkStorage** (covers R22). Measure get/list/subtree and current-open versus first-create on empty/100/1000 tasks plus 10000-task stress, with fixed note/tag/tree sizes. **Expect:** Record ns/op, allocations, compiler, OS/filesystem/CPU and raw benchmark output; setup is outside timed loops. No storage benchmark claims CLI <15 ms.

- [x] 002-V67 **Aggregate — StorageAcceptanceGates** (covers R1, R18, R21, R22). On the actual implementation revision, run minimum/current compiler proof, full/race/coverage, generated check, scripts and CGO-disabled target builds through Make. **Expect:** All applicable local gates pass with >=95% handwritten-package coverage; native/hosted evidence remains separate and pending if unexecuted.

- [x] 002-V68 **Operational handoff — StorageRunbookReplay** (covers R6, R7, R16, R21, R22). Follow docs/storage.md with disposable recognized/foreign/newer/failed fixtures and verify Phase 3 boundary references. **Expect:** Documented open/refusal/recovery and consistent-offline-backup guidance match behavior; triplet/masterplan evidence is synchronized and future hosted/native gates stay explicit.

## Commands and environments

| Command | Availability / owner | Evidence |
| --- | --- | --- |
| `make setup` | Exists; U6 raises version validation | Prerequisite check only; no module/tool download |
| `make test-unit` | Exists | Short tests; disk/process exclusions cannot establish full acceptance |
| `make test`, `make race` | Exist | All storage suites, including disk/process cases; record selected test names |
| `make validate` | Exists; U6 adds existing setup-script checks, U2 extends them | fmt/vet/full/race/coverage; >=95% handwritten nonexempt packages |
| `make test-compat` | Implemented U6; passes after 002-ISS-021 resolution | Driver/engine/transaction smoke plus fake minimum-version rejection |
| `make build-storage` | Implemented U6 | CGO-disabled native and five-target storage compilation |
| `make setup-sqlc` | Implemented U2 | Explicit pinned tool download/install when needed; digest checked |
| `make generate` | Implemented U2 | Scratch generation then replacement of only owned generated outputs |
| `make check-generated` | Implemented U2 | Non-mutating full-directory comparison, including untracked/obsolete files |
| `make test-scripts` | Implemented U2 | Setup/generator negative fixtures; no real network needed in negative tests |
| `make bench-storage` | Implemented U5 | Storage timing/allocations, reference environment and raw output |
| Native Windows/macOS / hosted jobs | Deferred Feature 006 | Exact candidate revision, job URLs, compiler, OS/arch and result; never inferred from cross-build |

The baseline Makefile has no supported TEST/PKG/RUN filter variables. New targets must be implemented and checked before invocation; until then use existing full/short targets. Each implementation unit records the first expected test failure before its production changes, then focused green and canonical aggregate results. Characterization cases already passing do not authorize unrelated changes.

## Fixtures and failure injection

- Allocate every file under a per-test temporary directory. Use no default home, actual TUSK_DB_PATH, or developer database. Inject environment/home/cwd for path tests.
- Seed an empty DB, a single root, independent roots, siblings with equal sort keys, a ten-level branch, NULL dates, nil/empty tags, Unicode strings, wildcard text, and metadata-only history. A 1000-task benchmark fixture uses 128-character notes, three tags, and one depth-10 branch; a 10000-task fixture is stress evidence.
- Use fixed UTC times with nanoseconds, zero/invalid/out-of-range timestamps, timezone offsets, and boundary due instants. Storage canonicalization does not test natural-language date parsing.
- Use unique named shared-memory DSNs per test with an anchor writer. Memory tests inspect MEMORY journal mode. Temp disk tests inspect WAL; never substitute one for the other.
- Migration fixtures use injected fs.FS inventory and disposable old/new/corrupt databases. LF-controlled production migration bytes and hashes are portable. Do not mutate real embedded source to provoke failure.
- A private test connector wrapper injects begin/statement/event/commit/rollback/close failures and captures operation order. It delegates real SQLite work where the behavior under test requires persistence. No runtime-global fault switches.
- Distinguish failure before commit reaches SQLite from failure after commit succeeds but its acknowledgment is lost. Readback is authoritative; error text alone cannot prove rollback.
- Use barriers, pipes, and bounded deadlines to place concurrent/process actions. No arbitrary sleep establishes ordering. A lock timeout test allows a recorded scheduling tolerance of at most 2 seconds beyond the 5-second busy limit. A 100-ms context must return before the full 5-second limit; U6 probes this before schema work, and U5 repeats it across processes. Native file-open and arbitrary filesystem I/O do not carry that cancellation guarantee.
- Add nil callback, ignored operation failure, panic, nested use, post-callback use, concurrent handle use, and Close admission cases. Callbacks must be cooperative and cannot invoke owner Close.
- Permission-bit denial tests must recognize privileged runners. Pair native POSIX mode checks with deterministic permission-error injection; Windows reparse/ACL tests require native execution.
- Corruption bypasses and sentinel secret-like title/note values exist only in fixtures. Error output must contain no raw user values; failed fixture retention must remain private and be cleaned after diagnosis.
- Kill helpers only after a specific pre-commit or post-ack handshake. Reap all children. A process kill leaves the OS running and cannot represent sudden hardware power loss.
- For decode failures or SQL row-iteration errors, assert no partial slice escapes and subsequent valid work can obtain a clean connection.

## Execution record

All six units are locally accepted. Native Windows/macOS execution remains pending; chronological receipts below retain earlier states.

| Date | Revision and dirty scope | Unit/scenario | Compiler and OS/arch | Command | Red/green/result | Evidence |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-09-08 | Base `4bbc639`, uncommitted U6 files on `feat/sqlite-storage-repository` | U6 initial red | Installed Go 1.27.1, linux/amd64 | `make test` | Expected compile failure | `compatibility_test.go`: `undefined: newConnector`, before production implementation |
| 2026-09-08 | Same U6 scope | 002-V02 red/green | Fake toolchains + installed Go, linux/amd64 | Temporary Make recipes `test-setup-red`, then `test-setup-proof`, running `scripts/test/test_scripts.sh` | Red: Go 1.24.9 incorrectly accepted; green: 9/9 script checks | Missing, malformed and 1.24 rejected; 1.25/newer accepted after setup change |
| 2026-09-08 | Same U6 scope | 002-V01/03/04/08, pool portion of V06 | Explicit `go1.25.0`, linux/amd64 | `GOTOOLCHAIN=go1.25.0 make test-compat` | Individual probes pass; target fails lock cancellation | Module assertions match sqlite v1.58.0/libc v1.75.6; engine 3.53.4; no-I/O construction, query-only reader, replacement pragmas and canceled pool acquisition pass. Deferred reader mode/replacement lock-mode proof is not yet complete. |
| 2026-09-08 | Same U6 scope | 002-V05/V06 lock probe | Go 1.25.0 / installed Go 1.27.1, linux/amd64 | Minimum and installed `make test-compat` | FAIL | `TestWriterConnection_BeginsImmediate`: 100-ms deadline returns `SQLITE_BUSY` after 5.03763051s / 5.009708107s; rollback then new writer succeeds. No early writer entry. |
| 2026-09-08 | Same U6 scope | U6 aggregate | Installed Go 1.27.1, linux/amd64 | `make validate`; `make race` | FAIL | fmt/vet pass; validate stops at full test failure (5.03194221s lock wait); separate race run fails same assertion (5.05168653s), no race report. Coverage gate not reached. |
| 2026-09-08 | Same U6 scope | 002-V07 | Explicit Go 1.25.0, linux/amd64 host | `GOTOOLCHAIN=go1.25.0 make build-storage` | PASS, exit 0 | CGO disabled: native storage build plus linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64. Compilation only; native execution is not proved. |

Record exact tests run and any short/native exclusions. A later source change invalidates affected proof and requires a focused rerun before aggregate closure. Hosted, native manual, publication, and local results use separate rows. Do not check a scenario merely because its test file exists.

### U6 resolution receipt (2026-09-08)

Observed red: the original 100-ms probe and new cancellation-without-deadline probe both returned SQLITE_BUSY after about 5 seconds. Green: original probe returns DeadlineExceeded after 102.319619 ms; cancellation without deadline and restoration to 5000 pass. Fault tests cover configure/begin/restore/rollback and incompatible/failed connections. Real SQLite tests cover five-second busy exhaustion, replacement deferred/query-only readers while a WAL writer holds its lock, and replacement writers acquiring after release. `GOTOOLCHAIN=go1.25.0 make test-compat build-storage` exits 0. Final `make validate` exits 0, including full/race, setup fixtures 9/9, and storage coverage 97.8%. U6 acceptance is local on base 4bbc639 plus the uncommitted U6 files; no native cross-target runtime or hosted proof.

### U1 execution receipt (2026-09-08)

Base 4bbc639 plus uncommitted U6/U1 changes. Red-first: missing db/Migrations package and migrator; missing disposable inverse fixture. Added regressions observed acceptance of foreign sqliteXsecret and a false incompatible-schema result when initialization committed between identity/catalog reads. Green fixes use literal catalog-prefix filtering and one inspection snapshot. TestMigrate_EmptyDatabase, TestMigrationInventory_Invalid/Unreadable/StableBytes, TestMigrate_FailurePreservesPreviousVersion, TestMigrate_RefusesForeignOrNewer/RefusalPreservesFile/BrokenLedger/CorruptLedgerAndCanceledInspection, TestInspection_FailureReturnsNoPartialCatalog, TestMigrate_InjectedFailures, TestMigrate_InspectionUsesOneSnapshot, TestMigrate_ConcurrentInitializers, TestSchema_RejectsInvalidRows, TestMigrate_MemoryAndInverseFixture and TestEmbed_ForwardOnly pass. Existing task data and schema roll back together; foreign main/WAL bytes remain unchanged; two helper processes initialize exactly one ledger entry; injected ledger/commit/rollback failures have old-or-new readback and successful recovery. Final make validate exits 0 (storage 96.1%, db 100% coverage), including race and process fixtures. ce-simplify-code reuse/quality/efficiency passes ran inline per repository policy: no behavior-preserving change warranted; canonical schema replay cost remains for the U5 benchmark. No publication/native cross-target execution claim.

### U2 execution receipt (2026-09-08)

Red: missing generated package and scripts/sqlc.sh; first real generation rejected ambiguous recursive ID references, corrected by qualification. Installed sqlc v1.31.1 archive matched the pinned linux/amd64 SHA-256. make generate and check-generated pass. TestQueries_CRUDAndCounts covers bound SQL-like text, nullable values, update/delete counts and metadata sequence/cascade; TestQueries_TreeSets covers depth ten, unrelated root, children and cyclic recursive-query termination; TestQueries_CandidateSuperset compares 112 enum/parent/date/search combinations to core. Script fixtures prove version/host/digest/truncated archive/extra member/path traversal/symlink/download refusal and prior-tool preservation; fake partial generation preserves prior output; whole-directory comparison detects stale/extra/missing output in a fixture with no Git metadata. A new red fixture exposed coverage.sh misreading Go output for packages without tests; the parser now honors the existing exact sqlc exemption while still rejecting a similarly named handwritten package at 0%. Final make validate check-generated passes: handwritten storage 96.1%, db 100%, script base 11/11 plus sqlc fixtures. All proof is local and uncommitted.

### U3 execution receipt (2026-09-08)

Red: undefined Open, Options, path inputs and connection roles. Local green: TestResolvePath_Precedence, TestOpen_LiteralFilename/RejectsUnsafeTarget/POSIXPermissions/ClosesPartialResources/FailureStagesReleaseEveryHandle/CanceledAndInvalidOverrideNeverFallsBack/ReplacementConnectionPragmas/MemoryLifetimeAndIsolation/CloseAdmissionAndIdempotency/CurrentSchemaNoMigrationWrite. These prove lazy fallback lookup, literal punctuation, symlink/FIFO refusal, restrictive new and preserved existing permissions, no fallback on invalid override, configured replacement pools, private memory anchors, admitted-work close behavior and current-schema opening during a held writer. Fault fixtures count every opened/closed physical handle across inspection connect/begin/catalog/close, WAL configuration, migration and reader ping failures; retry succeeds. make build-storage validate check-generated passes, storage coverage 95.7%. 002-V33 remains unchecked: the Windows test compiles, but native reparse/path/ACL proof belongs to Phase 6. U3 is locally accepted; final review remains pending.

### U4 execution receipt — 2026-09-08

Base 4bbc639 plus uncommitted implementation. make validate check-generated passed; handwritten storage coverage 97.0%. Go 1.25 test-compat and five CGO-disabled storage/test builds passed. Repository round-trip, exact core filtering, tree corruption, deletion, history, snapshot, lifetime/concurrent-handle, cancellation, commit/rollback uncertainty and driver-fault tests cover 002-V40–002-V58. Observed regressions before fixes: public Open leaked OS paths; ListChildren unnecessarily decoded corrupt grandchildren; rollback cleanup failures lost the original context/domain cause. Sanitized categories, immediate-child reads and joined safe causes resolve those failures. Callback replay remains prohibited, provisional failed reads return no data, and the read handle exposes no writer interface. U4 is locally accepted; U5 is active. Native/hosted execution is not claimed.

### U4 commit reconstruction — 2026-09-08

The original red-first work was accumulated without per-unit commits. At the user's correction, this unit was reconstructed in an isolated worktree and make validate was rerun on its exact code contents before committing. The original chronological test receipts above remain historical evidence. Unit completion now includes a separate local commit before advancing; pushing and merging are outside this authorization.

### U5 acceptance receipt — 2026-09-08

All applicable Feature 002 local scenarios are accepted: 67/68 checked, with only native Windows 002-V33 deferred to Phase 6. Current Go 1.27.1-X:nodwarf5 make validate check-generated build passed after the review fixes (storage coverage 97.6%, db/cmd 100%, core 98.2%; existing ports/sqlc exemptions unchanged). Explicit Go 1.25 previously passed the full gate; after the final mapper changes it again passed test-compat, all five CGO-disabled storage/test builds, check-generated and focused migration/inspection race tests. make setup and the final Btrfs make bench-storage run passed. No hosted or native macOS/Windows runtime result is claimed.

Two helper processes preserve all 40 read-modify-write increments and 40 events; a reader retains its snapshot across another process's commit. External writer cancellation returns in 102.5 ms (105.3 ms under race), while the uncanceled wait returns busy in 5.01 s; later writes succeed. Killing and reaping a writer before commit preserves the old task/event set; killing after acknowledgment preserves exactly one committed set. Independent reopen checks integrity and foreign keys. Held-reader WAL growth is followed by automatic restart sequence 0 -> 5 with bounded file reuse, without explicit checkpoint SQL. Normally closed offline backup reopens without modifying its source.

Observed red-first U5 fixes restore the default Make target, preserve migration statement/ledger/commit/rollback cancellation causes, preserve inspection cancellation, and retain unknown migration outcomes alongside schema categories. The completed ce-code-review receipt reported two actionable findings; both were reproduced and fixed, with no unapplied actionable residual. Review passes ran sequentially in the parent context per repository tool mapping; both independent peer routes failed before producing a review, so independent corroboration is unavailable. ce-simplify-code found no warranted behavior-preserving edit.

[Durable evidence, commit sequence and review resolution](../verification-evidence/002/README.md), [raw benchmarks](../verification-evidence/002/storage-benchmarks.txt), [code fingerprint and gate receipt](../verification-evidence/002/acceptance.json), and [operations/service handoff](../storage.md) retain the evidence. The separate U5 commit closes Phase 2; the next active target is Phase 3 planning, not Feature 003 implementation. No push, PR, merge or release is authorized by this acceptance.

### Post-acceptance publication follow-up — 2026-09-08

The user authorized simplify, review to zero actionable findings, compound, then commit/push/PR. MASTERPLAN target 2.4 tracks this follow-up; the six validated implementation commits remain separate. Simplification found no worthwhile behavior-preserving changes. Fresh review reports zero actionable findings; make validate check-generated build passes with 97.6% storage coverage. Both external review routes failed before producing usable review evidence; nine local passes ran sequentially in the parent context. See ../verification-evidence/002/publication-review.json. The reusable transaction-outcome/redaction lesson is captured in ../solutions/database-issues/preserve-transaction-outcomes-through-error-redaction.md with parser, link and source grounding checks. The branch is published as [PR #2](https://github.com/newbpydev/tusk/pull/2) against main; merge remains user-owned. Hosted checks and feedback are handled by the PR monitor, separately from local acceptance. Feature 003 implementation remains out of scope; V33 native Windows remains a Phase 6 obligation.
