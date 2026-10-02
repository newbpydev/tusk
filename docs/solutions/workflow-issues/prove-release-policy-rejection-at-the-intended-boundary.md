---
title: "Prove release-policy rejection at the intended boundary"
date: "2026-10-02"
category: "workflow-issues"
module: "Release tooling verification"
problem_type: "workflow_issue"
component: "development_workflow"
severity: "medium"
applies_when:
  - "Testing release checks with multiple refusal guards"
  - "Using fake platform tools or restricted PATH snapshots"
tags: ["release", "verification", "test-isolation", "fixtures", "red-green"]
---

# Prove release-policy rejection at the intended boundary

## Context

Feature 006 review reproduced malformed JSON acceptance, GNU-only fixture hashing
and a masked Windows path-conversion failure. Initial probes also exposed a
verification trap: the expected exit code can come from a different guard or an
incomplete fixture. Preserve those probes as investigation history, then replace
them with controlled evidence before accepting a fix.

## Guidance

1. Establish a passing control in an owned fixture with the same source,
   dependencies, selected binary and record identities as the negative case.
2. Change one rejection trigger. Clear independent remote fixture state and use
   a fresh output path so a prior tag, report or owned driver cannot decide the
   result first.
3. Keep every unrelated identity valid. After changing a hash-bound report,
   update its digest in acceptance, then update the acceptance digest in approval.
   Otherwise a hash mismatch can conceal the parser defect being tested.
4. Make fake tools obey the real invocation contract. Log or assert that the
   intended branch was reached; matching only the final exit code is insufficient.
5. Assert the expected refusal boundary and prohibited side effects. Use the
   diagnostic, call trace or filesystem evidence appropriate to that boundary.
   Keep real API reads distinct from privileged writes and from owned fake calls.
6. Observe Red before implementation, then Green with the same controlled case.
   Retain unsuccessful probes with their invalidating cause, and bind the final
   check to the source and environment actually used.

## Why This Matters

A refusal at a conflicting tag says nothing about approval parsing. An exit-zero
fixture that never entered Windows conversion says nothing about that converter.
These mistakes can either hide a defect or misidentify its cause. The control,
branch observation and side-effect assertion establish which behavior changed.

## When to Apply

Use this procedure for release approval, provenance, digest and native-platform
checks, or any negative regression with several earlier rejection guards. Fake
API and platform fixtures prove their modeled boundary; actual hosted and native
acceptance still needs the evidence required by the release plan.

## Examples

**Reset competing guards and repair dependent hashes.** In
`scripts/test/test_promotion.sh`, the concatenated-approval test resets remote
state before injecting a failed object followed by an approved one. The test
checks refusal and an empty owned fake API call log. The report variants at
`scripts/test/test_promotion.sh` recompute both dependent digests before
invocation. Their fake candidate verifier deliberately models a trusted boundary;
its empty call log does not imply production promotion performs no API reads.

**Honor fake-tool argument selection.** The Go fixture at
`scripts/test/test_release_smoke.sh` returns only each requested `go env` key.
The Windows case in that file supplies a MINGW
identity, a `cygpath` that exits 23, and a matching Windows manifest digest.
The corrected pre-fix run returned 0 where refusal 1 was required; the fixed run
refuses it. The earlier probe used a fake Go response that bypassed the intended
conversion path, so its matching failure was not causal evidence.

**Preserve dependencies in portability snapshots.** Remove only the tool under
test from an otherwise faithful source/PATH snapshot. The first smoke portability
probe omitted `go.mod`; that setup failure did not demonstrate a GNU hash
requirement. The corrected shasum-only snapshot reproduced failures in all four
policy targets, which subsequently passed with the portable fixture helper.
This Linux fixture is not a physical macOS acceptance run.

The [review receipt](../../verification-evidence/006/review-r2/acceptance.json)
identifies valid Red/Green checks and incomplete probes; the corresponding
compressed logs remain in that directory. The [portability comparison](../../verification-evidence/006/review-r2/checksum-portability-green.json)
lists all four target results.

**Preserve checkout bytes before testing digest integrity.** Native Windows CI
exposed approved dependency hash drift because `core.autocrlf` converted Markdown
fixtures. Keep first-party text at LF and pinned vendor files at `-text`. Put the
vendor rule after extension rules: an earlier `third_party/** -text` is overridden
by a later `*.go text`. The owned checkout regression retains a CRLF vendor Go
file and an LF first-party document under `autocrlf=true`; a fresh full checkout
also passes both real vendor inventories. Release archive fixtures copy the root
attributes and set their own `core.autocrlf=false` before adding files. Changing
approved hashes to match converted bytes would conceal the checkout defect.
See the [combined native CI repair receipt](../../verification-evidence/006/pr6-r2/acceptance.json).

**Match physical paths and every native consumer.** macOS temporary directories
can use `/var/...` while the driver's `pwd -P` resolves `/private/var/...`. A
literal argument assertion must use the same physical identity. Reproduce this
on Linux with an owned symlinked `TMPDIR`; label final assertions so a failed
path comparison is visible. On Windows, inspect the arguments sent to the
measurement harness as well as preflight's exported binary path. Pass converted
binary and report paths explicitly; conversion failure must occur before any
measurement launch. An argument fixture proves this boundary without claiming
native Windows execution or depending on implicit shell argument conversion.
The [third complete-batch receipt](../../verification-evidence/006/pr6-r3/acceptance.json)
retains that Red/Green comparison and the still-pending native runtime limits.

**Keep restricted tool snapshots faithful.** [Git Bash's default `ln -s`
creates copies](https://gitforwindows.org/symbolic-links.html). Copying an
executable into a restricted PATH can separate it from its adjacent runtime;
[Windows searches the executable directory for DLLs](https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-search-order).
Use Bash launchers that execute each installed image at its original location.
Keep the restricted PATH when testing prerequisite discovery; a backend-only
hash launcher may restore its own dependencies after the helper selects it.
An owned resolved-image control passes, while the same snapshot under copy
semantics fails. This models the suspected native loader cause; the subsequent
Windows gate must confirm execution.

Link emulation can also turn a symlink-archive fixture into a regular file. An
empty copied payload then fails at extraction size, concealing the absent
non-regular-member refusal. Keep the companion nonempty and assert the exact
refusal diagnostic. The deterministic `scripts/test/fixtures/sqlc-symlink.tar.gz`
contains an actual symlink header on every host without filesystem privileges.
See the [fourth complete-batch receipt](../../verification-evidence/006/pr6-r4/acceptance.json).

**Confirm the fake compiler was selected.** Native Windows GNU Make 4.4.1
[searches extensions across PATH before extensionless files](https://github.com/mirror/make/blob/4.4.1/src/w32/subproc/sub_proc.c#L479-L527).
A simple `go test` recipe can therefore find the installed `go.exe` before an
owned shell-script `go`, although Bash itself selects the fixture. Force shell
lookup only for Make calls that inject those fake tools:
`make '.SHELLFLAGS=-e -c' ...`. The separate flags bypass Make's direct-command
optimization; the regular canonical recipes and compiler selection stay intact.
Require the fixture's call trace and its `Error 19` diagnostic alongside Make's
exit 2. A missing trace identifies selection failure; an unrelated Make error
cannot satisfy the negative test. Include recursive catalog generation in the
same check, because its `bash` and `go` fixtures share this lookup boundary.
The [fifth complete-batch receipt](../../verification-evidence/006/pr6-r5/acceptance.json)
keeps the native failure separate from local confirmation and fresh hosted proof.

**Compare paths after the native argument boundary.** [MSYS converts POSIX-looking
arguments and environment values when launching native Windows tools](https://www.msys2.org/docs/filesystem-paths/).
A Make dry run can correctly print `C:/...` although Bash's input was `/tmp/...`.
Use `cygpath -m` for the profile fixture's selected path on MINGW/MSYS, then assert
both the printed recipe and the actual compiler's `-cpuprofile`/`-o` arguments.
Keep the path with spaces and a fresh call trace. A controlled conversion model
reproduces the old lexical comparison failure and passes after explicit conversion;
it does not replace the subsequent native Windows run. See the
[sixth complete-batch receipt](../../verification-evidence/006/pr6-r6/acceptance.json).

**Keep fixture resource budgets separate from performance assertions.** An owned
dependency checkout shared one ten-second context across Git init, staging and
checkout. Two Windows attempts stopped during staging after roughly ten seconds;
unchanged earlier attempts passed. A controlled five-second delay before each
real Git operation reproduces the coupled-budget failure and then completes the
same full checkout after repair. Give each operation a bounded 30-second context
and cancel it immediately afterward. This changes the fixture's resource budget,
while its required failing LFS filter, pointer refusal and product latency targets
remain intact. [CommandContext interrupts its child when the context ends](https://pkg.go.dev/os/exec#CommandContext);
record `ctx.Err()` before calling `cancel()` so an actual expiry is distinguishable
from a tool error. A stalled owned process still fails near 30 seconds with the
explicit deadline diagnostic. Keep the original retry failure and the initial
transient-failure hypothesis in the [seventh complete-batch receipt](../../verification-evidence/006/pr6-r7/acceptance.json).

**Choose an explicit mismatch for negative identity tests.** Requesting
`windows/amd64` from the ambient compiler is a negative case only on other hosts;
it becomes a valid positive control on Windows. Use the owned `windows/amd64`
identity for both cases, then request `windows/arm64` for the negative. Require
the exact OS/architecture guard diagnostic as well as exit 1. The controlled
Windows-identity shell suite reproduces the old false failure and confirms the
repair; subsequent native CI still supplies the actual host proof. See the
[eighth complete-batch receipt](../../verification-evidence/006/pr6-r8/acceptance.json).

**Exercise both operands of native identity guards.** The owned windows/amd64
control can reject a requested linux/amd64 OS independently of the existing
windows/arm64 architecture negative. Both cases require the exact job-identity
diagnostic. An owned mutation that removes only the OS comparison escaped the
old suite and is rejected after adding the dedicated negative. The production
identity predicate itself does not change; no extra fake-platform knob is needed.

**Select LF at the JSON producer.** [Native Windows jq translates output LF to
CRLF unless binary output is selected](https://jqlang.org/manual/v1.7/#invoking-jq).
A notices control then compares LF filesystem paths with CRLF JSON-derived paths
and reports an incomplete inventory despite identical asset names. Select
`jq --binary` on MSYS/Cygwin for the maintainer helpers, their fixtures and
the Make compiler-pin lookup; leave Unix invocations unflagged. This keeps paths, TSV fields and newly generated records consistent;
it does not remove carriage returns from source license text or JSON values.
Require jq 1.7+ with binary output support on Windows and probe that capability
at setup/preflight. Older Unix jq builds reject this Windows option; presence
alone does not establish Windows support. A permanent notices fixture
models this translation and preserves jq exit codes; an owned full-shell model
checks the other callers. Its Linux bootstrap failure is a broader model result,
separate from the actual Windows notices failure. Fresh native execution remains
required to confirm the native line-ending hypothesis. See the
[ninth complete-batch receipt](../../verification-evidence/006/pr6-r9/acceptance.json).

**Use native launchers for native process fixtures.** Go's Windows executable
lookup can select installed `gh.exe` ahead of an extensionless Bash fake. Build
an owned native `gh.exe` that dispatches the owned Bash script with literal argv,
forwarded streams and the original exit status. Use the same launcher for
candidate, Homebrew, promotion and metadata tests so native API and mutation fixtures
cannot reach the installed client. A missing launcher configuration fails closed.
Require named assertions for argv/streams/status, map a Unix signaled child to
128 plus its signal number, and use a native fixture stall for deadline controls
so cancellation does not leave a Bash grandchild. Record the host jq identity
when retaining a CRLF model that depends on the host binary-output capability.
The test launcher is excluded from product builds; the production deadline and
GitHub client remain unchanged. Retain actual Windows evidence separately from
local launcher controls. See the
[tenth complete-batch receipt](../../verification-evidence/006/pr6-r10/acceptance.json).

**Synchronize current headers as well as appended receipts.** Historical sections
can be correct while top-level status still names an older review batch. Update
the master status, active pointer and paired plan/workorder evidence headers in
the same unit, keeping local closures separate from pending native/release gates.

**Cancellation is retried until the reader joins.** A successful cancellation
request is not proof the reader has returned. The confirmation test's fake
canceller must tolerate another retry after unblocking its channel. Hold the
fake reader until a third request, close each synchronization channel only once,
and continue requiring cancellation plus completed thread release. This
reproduces the prior double-close panic deterministically without changing the
production retry/join behavior. The
[eleventh complete-batch receipt](../../verification-evidence/006/pr6-r11/acceptance.json)
retains the native macOS failure and local Red/Green control.

## Related

- [Release acceptance and publication procedure](../../releasing.md)
- [Owned Kitty verification](verify-owned-kitty-window-without-focus-or-environment-leaks.md)
  applies the same evidence-scope discipline to visible terminal checks.
