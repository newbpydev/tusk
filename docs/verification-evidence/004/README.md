# Feature 004 execution evidence

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
