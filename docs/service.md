# Task service

`internal/ports.TaskService` is the synchronous application API shared by CLI,
TUI and automation. `service.NewTaskService` requires a repository, a clock, an
ID function and a location. Construction performs no I/O. The caller owns the
repository and closes it after all operations finish. Injected functions must
support concurrent callers. Each operation owns its state; there is no cache,
worker, global pool or service lock.

Use `service.NewUUIDv7(reference, entropy)` with a cryptographic entropy source
at the application composition root. IDs are lowercase UUIDv7 and disclose a
millisecond creation time. They are not credentials. Duplicate IDs fail without
retry, including collisions with tasks already loaded for hierarchy work.

Mutations preflight detached commands, capture one UTC reference time, and use
one writer callback. Supplied fields patch the latest task; omitted fields remain
unchanged. An optional Base compares editable values and CreatedAt, excluding
derived parent progress. It accepts a value changed and restored to equality.
Empty/equal patches preserve timestamps and history. Manual progress is for
leaves and cannot accompany parent/status intent.

Complete affects the full subtree. Reopen preserves descendants, resets a leaf
to zero, and recalculates parents. An explicitly reopened parent can remain open
at 100. AutoCompleteParent defaults false; when enabled it requires every direct
child to be done, rather than merely at 100. Moves validate cycles and final
subtree depth under the writer lock. Last-child removal resets an open parent
to zero; an untouched leaf keeps its manual progress.

Delete first uses PreviewDeleteTask to capture the target and exact sorted subtree
IDs. The adapter displays scope and asks for confirmation; pass that preview as
Expected. Changed membership/metadata fails with ErrConflict. Force omits the
preview check, but never implies Recursive. Deleted task history is removed; no
tombstone or undo is retained. Preview is a snapshot, not an authorization token.

Lists default to open statuses and sort by priority descending, due ascending
(null last), creation ascending and ID ascending. Due-day predicates are
half-open local civil days. Trees include done tasks and preserve stored parent
IDs even when a selected subtree is displayed at depth 1. Stats count all
retained tasks, including parents. CompletedLast7Days counts currently done tasks
in (reference-168h, reference]; it is not immutable productivity history.

All returned values are detached. Errors return nil task/list/tree/history or
zero stats/preview/delete results. Known commit success remains success if the
context is canceled after commit. History contains only sorted changed field
names, never prior/new title or note values. Stored Markdown and terminal control
bytes remain unchanged; adapters own escaping and sanitization for display.

## Recovery from an unknown outcome

Check errors.As for ports.TransactionError before interpreting a cause as
retryable. An unknown outcome can still match cancellation, conflict or storage
categories. Do not repeat the callback or present the staged result as success.
Stop using that repository, close it, open a new owner for the same database and
read tasks/history. The state may be wholly old or wholly new; only fresh
readback determines it. A read cleanup failure also discards assembled results.
The service does not close shared resources or recover automatically.

The executable runbook is TestService_RollbackAndUnknownOutcome and
TestService_CommittedThenCanceled, using temporary disk databases. Process
termination fixtures exercise before-first-write, partial-write and
acknowledged-commit barriers, always reaping the child and verifying integrity
and foreign keys without deleting WAL/SHM files. These are process-failure tests,
not hardware power-loss proof.

## Dates and consumer handoffs

The parser accepts today/tomorrow/tonight, three-letter weekdays, positive
calendar day/week/month offsets, ISO dates and strict offset timestamps. Month
arithmetic clamps the final destination month. Missing wall times or midnight
boundaries fail; repeated wall times select the earlier instant. Parsing uses
only the supplied reference/location and returns UTC. Test binaries embed
`time/tzdata`; Feature 004 must embed it in the production composition root.

Feature 004 owns CLI grammar, JSON DTOs, exit codes, policy/zone configuration,
output-after-commit behavior and end-to-end startup/query latency. Feature 005
owns consent UI, draft preservation, refresh generations, terminal accessibility
and escaping. Feature 006 owns native Windows/macOS execution and hosted release
proof. Cross-compilation does not close those gates.

`make validate` is the local quality gate. `GOTOOLCHAIN=go1.25.0 make test
build-service` verifies the minimum compiler and compiles service/parser tests
for five CGO-free targets. `make bench-service` records one warm sample for each
workload/size, with setup/reset outside timing. It reports allocations, repository
read/write call counts and snapshot duration. Those call counts describe port
operations, not SQLite virtual-machine statement counts. Single samples are
baseline observations, not statistical latency guarantees or CLI acceptance.
See [Feature 003 evidence](verification-evidence/003/README.md).
