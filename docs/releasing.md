# Release procedure and public repository preview

No public release exists. The first proposed version is `v0.3.0`; the owner must
approve the actual version and source commit. Local engineering and commits do
not authorize pushing, workflow dispatch, tags, settings changes or publication.
Binary commands in [installation](install.md#binary-download-drafts) are draft
instructions until real URLs and native receipts exist.

## Candidate workflow

Keep Go 1.25 as the source floor; production builds use the compiler, GoReleaser
and action hashes pinned in `scripts/tool-versions.json`. Use the complete source
checkout with replacements and notices. Never run an ambient packager as proof
of the pinned release tool.

Before a candidate, run canonical validation and generated/module/license checks.
Local packaging creates immutable version input and five CGO-free archives,
source, notices, checksums and a manifest. The intended archive format is
`tusk_VERSION_OS_ARCH.tar.gz` on Unix and `.zip` on Windows, with root `tusk` or
`tusk.exe`, README, LICENSE, THIRD_PARTY_NOTICES, completions and manpages.
Linux ELF inspection must prove no interpreter/shared-library dependency; Darwin
and Windows may use normal OS libraries. Avoid a universal “fully static” claim.

Use Bash and GNU Make with `jq`, Git, `curl` and the pinned Go toolchain available.
Install the verified packager once, then check the current source prerequisites:

```bash
make setup-sqlc setup-release
make test-release release-check
```

Select a full commit SHA containing the release tooling. Use a new output directory
whose parent exists, with a separate directory for each retained run:

```bash
make release-candidate RELEASE_VERSION=v0.3.0 RELEASE_SHA=FULL_COMMIT_SHA RELEASE_OUTPUT=/tmp/tusk-candidate-unique
make verify-release RELEASE_ASSETS=/tmp/tusk-candidate-unique/assets
make release-snapshot RELEASE_VERSION=v0.3.0 RELEASE_SHA=FULL_COMMIT_SHA RELEASE_OUTPUT=/tmp/tusk-snapshot-unique
```

`FULL_COMMIT_SHA` and the output names are placeholders. Candidate binaries report
`0.3.0`; snapshots report `0.3.0-dev`; ordinary source builds report `dev`. Local
candidate names express intended version identity and do not prove publication.
The helper creates a detached temporary checkout and local tag, runs canonical
release prerequisites there, applies a generated constant overlay outside the checkout, and
invokes GoReleaser with publication disabled and an environment without publishing
credentials. The caller's source, tags, tasks and user configuration are preserved.
An existing conflicting tag or retained output path is rejected.

Git source-archive modes use a fixed 0022 tar umask and tracked executable bits.
GoReleaser's `dist` setting is literal, so the helper generates an external runtime
config changing only that field to owned temporary storage. The canonical config
hash remains stable in the manifest; `config-input.json` retains the actual path
override as per-run diagnostics.

The `assets` directory contains exactly five platform archives, source, standalone
notices, checksums and a manifest. The manifest binds source, compiler, tool/config,
module/license/patch inputs, generated constant, executable and member hashes.
Checksums cover the seven payload/notice files; the manifest records their hashes
and the checksum-file hash, avoiding a self-hash cycle. Hosted run/provenance
identity belongs in a separate receipt. Binary archives include guides and the
sample screenshot; source archives include the complete engineering documentation
and patched module sources. Relative engineering links in the contribution guides
can be followed in the source archive or repository.

`build.log`, `version-input.json`, `config-input.json` and `result.json` remain beside successful assets.
Failed runs keep the available diagnostics and never replace an earlier run. Only
the invocation-created temporary build storage is cleaned. Compare executable and
archive hashes from two locations when accepting reproducibility; do not infer it
from stable filenames.

The hosted candidate workflow is not implemented/activated yet. Its reviewed
contract is a trusted exact main SHA, no publishing/tap secrets and attested
payload/manifest hashes. Once authorized, build once and download those exact
bytes for native install, configuration, CLI/TUI, backup, upgrade, uninstall and
retained performance checks. A cross-build is not native execution. Required
hosts are Linux amd64/arm64, macOS Intel/ARM and native Windows amd64, including
an owned Windows 11 terminal. Missing access leaves the gate pending.

## Publication and failure recovery

Owner authorization must name version/SHA and accepted manifest/run before
creating a draft. Verify every asset, checksum, notice and provenance record in
the draft, then publish only after concrete final authorization. Promote accepted
files without rebuilding. No automatic public release occurs on push/PR.
Do not overwrite published bytes or retag. Reconcile API state after a lost
response before retrying; deprecate a faulty release and fix it under a new version.

The macOS cask destination proposal is `newbpydev/homebrew-tap`, `Casks/tusk.rb`.
The owner must establish destination/access. Validate Intel/ARM hashes, binary,
manuals and completions locally, then publish the reviewed cask after release
URLs work. Cask removal must preserve task data; no quarantine-removal hook.
A failed tap update leaves the phase pending. No tap push is authorized here.

## Proposed GitHub metadata

Apply these values only after owner authorization and retain before/after readback:

- Description: `Local task management in your terminal: a keyboard-driven TUI and scriptable CLI, backed by SQLite.`
- Homepage: `https://github.com/newbpydev/tusk#readme`
- Topics: `go`, `golang`, `cli`, `tui`, `task-manager`, `terminal`, `sqlite`,
  `bubbletea`, `productivity`, `offline`, `command-line`.
- Social preview: the inspected sample capture `docs/assets/tusk-tui.png`, with
  provenance in [assets](assets/README.md). Use manual upload if needed.
- Website/funding: leave absent optional values empty; no separate website exists.
- Security: private vulnerability reporting is currently disabled; do not claim
  it is active without verified setting readback and an updated policy.

After publication, verify anonymous downloads/install, complete release assets,
Homebrew, GitHub About/license/topics/homepage/social preview and actual rendered
README. Activate badges only when their real CI/release/license targets resolve.
Recheck the README's first-task examples against the published files. Preserve
native/hosted/publication evidence separately from local source checks.
